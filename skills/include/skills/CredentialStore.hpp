#pragma once

#include <QMap>
#include <QObject>
#include <QString>
#include <QStringList>
#include <QVector>
#include <QSqlDatabase>

namespace skills {

struct ServiceCredential {
    QString clientId;
    QString clientSecret;
    QString refreshToken;
    QString region;
    QString environment;
};

struct ClientProfile {
    QString displayName;
    QMap<QString, ServiceCredential> serviceCredentials; // service name -> credential
};

class CredentialStore : public QObject {
    Q_OBJECT
public:
    explicit CredentialStore(QObject *parent = nullptr);

    ~CredentialStore();

    void clear();
    bool loadFromFile(const QString &filePath, QString *error = nullptr);

    QVector<ClientProfile> clients() const;
    const ClientProfile *findClient(const QString &name) const;
    QStringList servicesForClient(const QString &name) const;

    bool addOrUpdateClient(const ClientProfile &profile, QString *error = nullptr);
    bool removeClient(const QString &name, QString *error = nullptr);

signals:
    void storeChanged();

private:
    bool ensureSchema(QString *error);
    bool reloadFromDatabase(QString *error);
    void close();

    QVector<ClientProfile> m_clients;
    QString m_databasePath;
    QString m_connectionName;
    QSqlDatabase m_database;
};

} // namespace skills
