#include "skills/CredentialStore.hpp"

#include <QDir>
#include <QFileInfo>
#include <QMap>
#include <QSqlError>
#include <QSqlQuery>
#include <QStringList>
#include <QVariant>
#include <QtGlobal>

namespace skills {

CredentialStore::CredentialStore(QObject *parent)
    : QObject(parent)
    , m_connectionName(QStringLiteral("credential-store-%1").arg(quintptr(this), 0, 16))
{
}

CredentialStore::~CredentialStore()
{
    close();
}

void CredentialStore::close()
{
    if (m_database.isValid()) {
        m_database.close();
        const QString connectionName = m_connectionName;
        m_database = QSqlDatabase();
        QSqlDatabase::removeDatabase(connectionName);
    }
    m_databasePath.clear();
}

void CredentialStore::clear()
{
    if (!m_clients.isEmpty()) {
        m_clients.clear();
        emit storeChanged();
    }
}

bool CredentialStore::loadFromFile(const QString &filePath, QString *error)
{
    close();

    if (filePath.trimmed().isEmpty()) {
        if (error) {
            *error = tr("Credential database path is empty");
        }
        return false;
    }

    QFileInfo info(filePath);
    QDir directory = info.dir();
    if (!directory.exists()) {
        if (!directory.mkpath(QStringLiteral("."))) {
            if (error) {
                *error = tr("Unable to create directory for credential database: %1")
                              .arg(directory.absolutePath());
            }
            return false;
        }
    }

    m_database = QSqlDatabase::addDatabase(QStringLiteral("QSQLITE"), m_connectionName);
    m_database.setDatabaseName(info.absoluteFilePath());

    if (!m_database.open()) {
        if (error) {
            *error = tr("Unable to open credential database: %1").arg(m_database.lastError().text());
        }
        close();
        return false;
    }

    if (!ensureSchema(error)) {
        close();
        return false;
    }

    m_databasePath = info.absoluteFilePath();
    return reloadFromDatabase(error);
}

QVector<ClientProfile> CredentialStore::clients() const
{
    return m_clients;
}

const ClientProfile *CredentialStore::findClient(const QString &name) const
{
    for (const auto &client : m_clients) {
        if (client.displayName.compare(name, Qt::CaseInsensitive) == 0) {
            return &client;
        }
    }
    return nullptr;
}

QStringList CredentialStore::servicesForClient(const QString &name) const
{
    const auto *client = findClient(name);
    if (!client) {
        return {};
    }
    return client->serviceCredentials.keys();
}

bool CredentialStore::addOrUpdateClient(const ClientProfile &profile, QString *error)
{
    if (!m_database.isValid() || !m_database.isOpen()) {
        if (error) {
            *error = tr("Credential database is not open");
        }
        return false;
    }

    const QString trimmedName = profile.displayName.trimmed();
    if (trimmedName.isEmpty()) {
        if (error) {
            *error = tr("Client name cannot be empty");
        }
        return false;
    }

    if (!m_database.transaction()) {
        if (error) {
            *error = tr("Unable to start transaction: %1").arg(m_database.lastError().text());
        }
        return false;
    }

    QSqlQuery removeQuery(m_database);
    removeQuery.prepare(QStringLiteral("DELETE FROM credentials WHERE client_name = ?"));
    removeQuery.addBindValue(trimmedName);
    if (!removeQuery.exec()) {
        if (error) {
            *error = tr("Failed to clear existing credentials: %1").arg(removeQuery.lastError().text());
        }
        m_database.rollback();
        return false;
    }

    for (auto it = profile.serviceCredentials.cbegin(); it != profile.serviceCredentials.cend(); ++it) {
        QSqlQuery insertQuery(m_database);
        insertQuery.prepare(QStringLiteral(
            "INSERT INTO credentials (client_name, service_name, client_id, client_secret, refresh_token, region, environment)"
            " VALUES (?, ?, ?, ?, ?, ?, ?)"));
        insertQuery.addBindValue(trimmedName);
        insertQuery.addBindValue(it.key());
        insertQuery.addBindValue(it.value().clientId);
        insertQuery.addBindValue(it.value().clientSecret);
        insertQuery.addBindValue(it.value().refreshToken);
        insertQuery.addBindValue(it.value().region);
        insertQuery.addBindValue(it.value().environment);
        if (!insertQuery.exec()) {
            if (error) {
                *error = tr("Failed to store %1 credentials: %2")
                                 .arg(it.key(), insertQuery.lastError().text());
            }
            m_database.rollback();
            return false;
        }
    }

    if (!m_database.commit()) {
        if (error) {
            *error = tr("Unable to commit credential changes: %1").arg(m_database.lastError().text());
        }
        m_database.rollback();
        return false;
    }

    return reloadFromDatabase(error);
}

bool CredentialStore::removeClient(const QString &name, QString *error)
{
    if (!m_database.isValid() || !m_database.isOpen()) {
        if (error) {
            *error = tr("Credential database is not open");
        }
        return false;
    }

    if (!m_database.transaction()) {
        if (error) {
            *error = tr("Unable to start transaction: %1").arg(m_database.lastError().text());
        }
        return false;
    }

    QSqlQuery query(m_database);
    query.prepare(QStringLiteral("DELETE FROM credentials WHERE client_name = ?"));
    query.addBindValue(name);
    if (!query.exec()) {
        if (error) {
            *error = tr("Failed to remove client: %1").arg(query.lastError().text());
        }
        m_database.rollback();
        return false;
    }

    const bool removed = query.numRowsAffected() > 0;

    if (!m_database.commit()) {
        if (error) {
            *error = tr("Unable to commit removal: %1").arg(m_database.lastError().text());
        }
        m_database.rollback();
        return false;
    }

    if (!reloadFromDatabase(error)) {
        return false;
    }

    return removed;
}

bool CredentialStore::ensureSchema(QString *error)
{
    if (!m_database.isValid()) {
        if (error) {
            *error = tr("Credential database is not initialised");
        }
        return false;
    }

    QSqlQuery pragma(m_database);
    pragma.exec(QStringLiteral("PRAGMA foreign_keys = ON"));

    QSqlQuery createQuery(m_database);
    const QString statement = QStringLiteral(
        "CREATE TABLE IF NOT EXISTS credentials ("
        " client_name TEXT NOT NULL,"
        " service_name TEXT NOT NULL,"
        " client_id TEXT,"
        " client_secret TEXT,"
        " refresh_token TEXT,"
        " region TEXT,"
        " environment TEXT,"
        " PRIMARY KEY(client_name, service_name)"
        ")");
    if (!createQuery.exec(statement)) {
        if (error) {
            *error = tr("Failed to ensure credential schema: %1").arg(createQuery.lastError().text());
        }
        return false;
    }

    return true;
}

bool CredentialStore::reloadFromDatabase(QString *error)
{
    if (!m_database.isValid() || !m_database.isOpen()) {
        clear();
        return true;
    }

    QSqlQuery query(m_database);
    if (!query.exec(QStringLiteral(
            "SELECT client_name, service_name, client_id, client_secret, refresh_token, region, environment "
            "FROM credentials ORDER BY LOWER(client_name), LOWER(service_name)"))) {
        if (error) {
            *error = tr("Failed to read credential data: %1").arg(query.lastError().text());
        }
        return false;
    }

    QMap<QString, ClientProfile> profiles;
    while (query.next()) {
        const QString clientName = query.value(0).toString();
        const QString serviceName = query.value(1).toString();

        if (clientName.isEmpty()) {
            continue;
        }

        ClientProfile &profile = profiles[clientName];
        profile.displayName = clientName;

        if (!serviceName.isEmpty()) {
            ServiceCredential credential;
            credential.clientId = query.value(2).toString();
            credential.clientSecret = query.value(3).toString();
            credential.refreshToken = query.value(4).toString();
            credential.region = query.value(5).toString();
            credential.environment = query.value(6).toString();
            profile.serviceCredentials.insert(serviceName, credential);
        }
    }

    QVector<ClientProfile> refreshed;
    refreshed.reserve(profiles.size());
    for (auto it = profiles.cbegin(); it != profiles.cend(); ++it) {
        refreshed.push_back(it.value());
    }

    m_clients = refreshed;
    emit storeChanged();
    return true;
}

} // namespace skills
