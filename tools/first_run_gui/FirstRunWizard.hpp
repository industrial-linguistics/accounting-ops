#pragma once

#include "skills/CredentialStore.hpp"

#include <QVector>
#include <QWizard>

class QCheckBox;
class QLineEdit;
class QLabel;
class QPushButton;

class ClientInfoPage : public QWizardPage {
    Q_OBJECT
public:
    explicit ClientInfoPage(QWidget *parent = nullptr);

    QString clientName() const;

private:
    QLineEdit *m_clientNameEdit;
};

class ServiceCredentialPage : public QWizardPage {
    Q_OBJECT
public:
    ServiceCredentialPage(const QString &serviceKey, const QString &serviceName, QWidget *parent = nullptr);

    bool isConfigured() const;
    QString serviceKey() const;
    skills::ServiceCredential credential() const;

protected:
    bool validatePage() override;

private slots:
    void handleTestClicked();
    void handleEnableToggled(bool enabled);

private:
    QString m_serviceKey;
    QString m_serviceName;
    QCheckBox *m_enableBox;
    QLineEdit *m_clientIdEdit;
    QLineEdit *m_clientSecretEdit;
    QLineEdit *m_refreshTokenEdit;
    QLineEdit *m_regionEdit;
    QLineEdit *m_environmentEdit;
    QLabel *m_statusLabel;
    QPushButton *m_testButton;
};

class FirstRunWizard : public QWizard {
    Q_OBJECT
public:
    explicit FirstRunWizard(skills::CredentialStore *store, QWidget *parent = nullptr);

protected:
    void accept() override;

private:
    skills::CredentialStore *m_store;
    ClientInfoPage *m_clientPage;
    QVector<ServiceCredentialPage *> m_servicePages;
};
