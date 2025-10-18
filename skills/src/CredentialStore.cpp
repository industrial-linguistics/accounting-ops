#include "skills/CredentialStore.hpp"

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>

namespace skills {

CredentialStore::CredentialStore(QObject *parent)
    : QObject(parent) {}

void CredentialStore::clear() {
    if (!m_clients.isEmpty()) {
        m_clients.clear();
        emit storeChanged();
    }
}

bool CredentialStore::loadFromFile(const QString &filePath, QString *error) {
    QFile file(filePath);
    if (!file.exists()) {
        if (error) {
            *error = tr("Credential file does not exist: %1").arg(filePath);
        }
        return false;
    }
    if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
        if (error) {
            *error = tr("Unable to open credential file: %1").arg(file.errorString());
        }
        return false;
    }

    const auto data = file.readAll();
    const auto doc = QJsonDocument::fromJson(data);
    if (!doc.isObject()) {
        if (error) {
            *error = tr("Credential file is not a JSON object");
        }
        return false;
    }

    const auto root = doc.object();
    const auto clientsArray = root.value("clients").toArray();

    QVector<ClientProfile> parsed;
    parsed.reserve(clientsArray.size());

    for (const auto &clientValue : clientsArray) {
        const auto clientObj = clientValue.toObject();
        ClientProfile profile;
        profile.displayName = clientObj.value("name").toString();
        const auto services = clientObj.value("services").toObject();
        for (auto it = services.begin(); it != services.end(); ++it) {
            profile.serviceCredentials.insert(it.key(), credentialFromJson(it.value().toObject()));
        }
        if (!profile.displayName.isEmpty()) {
            parsed.push_back(profile);
        }
    }

    m_clients = parsed;
    emit storeChanged();
    return true;
}

bool CredentialStore::saveToFile(const QString &filePath, QString *error) const {
    QJsonArray clientsArray;
    for (const auto &client : m_clients) {
        QJsonObject clientObj;
        clientObj.insert("name", client.displayName);
        QJsonObject servicesObj;
        for (auto it = client.serviceCredentials.cbegin(); it != client.serviceCredentials.cend(); ++it) {
            servicesObj.insert(it.key(), credentialToJson(it.value()));
        }
        clientObj.insert("services", servicesObj);
        clientsArray.push_back(clientObj);
    }

    QJsonObject root;
    root.insert("clients", clientsArray);

    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text | QIODevice::Truncate)) {
        if (error) {
            *error = tr("Unable to write credential file: %1").arg(file.errorString());
        }
        return false;
    }

    QJsonDocument doc(root);
    file.write(doc.toJson(QJsonDocument::Indented));
    file.close();
    return true;
}

QVector<ClientProfile> CredentialStore::clients() const {
    return m_clients;
}

const ClientProfile *CredentialStore::findClient(const QString &name) const {
    for (const auto &client : m_clients) {
        if (client.displayName.compare(name, Qt::CaseInsensitive) == 0) {
            return &client;
        }
    }
    return nullptr;
}

QStringList CredentialStore::servicesForClient(const QString &name) const {
    const auto *client = findClient(name);
    if (!client) {
        return {};
    }
    return client->serviceCredentials.keys();
}

void CredentialStore::addOrUpdateClient(const ClientProfile &profile) {
    for (auto &client : m_clients) {
        if (client.displayName.compare(profile.displayName, Qt::CaseInsensitive) == 0) {
            client = profile;
            emit storeChanged();
            return;
        }
    }
    m_clients.push_back(profile);
    emit storeChanged();
}

bool CredentialStore::removeClient(const QString &name) {
    for (int i = 0; i < m_clients.size(); ++i) {
        if (m_clients[i].displayName.compare(name, Qt::CaseInsensitive) == 0) {
            m_clients.removeAt(i);
            emit storeChanged();
            return true;
        }
    }
    return false;
}

QJsonObject CredentialStore::credentialToJson(const ServiceCredential &credential) {
    QJsonObject object;
    object.insert("clientId", credential.clientId);
    object.insert("clientSecret", credential.clientSecret);
    object.insert("refreshToken", credential.refreshToken);
    object.insert("region", credential.region);
    object.insert("environment", credential.environment);
    return object;
}

ServiceCredential CredentialStore::credentialFromJson(const QJsonObject &object) {
    ServiceCredential credential;
    credential.clientId = object.value("clientId").toString();
    credential.clientSecret = object.value("clientSecret").toString();
    credential.refreshToken = object.value("refreshToken").toString();
    credential.region = object.value("region").toString();
    credential.environment = object.value("environment").toString();
    return credential;
}

} // namespace skills
