#pragma once

#include <QJsonObject>
#include <QMap>
#include <QObject>
#include <QString>
#include <QStringList>
#include <QVector>

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

    void clear();
    bool loadFromFile(const QString &filePath, QString *error = nullptr);
    bool saveToFile(const QString &filePath, QString *error = nullptr) const;

    QVector<ClientProfile> clients() const;
    const ClientProfile *findClient(const QString &name) const;
    QStringList servicesForClient(const QString &name) const;

    void addOrUpdateClient(const ClientProfile &profile);
    bool removeClient(const QString &name);

signals:
    void storeChanged();

private:
    QVector<ClientProfile> m_clients;

    static QJsonObject credentialToJson(const ServiceCredential &credential);
    static ServiceCredential credentialFromJson(const QJsonObject &object);
};

} // namespace skills
