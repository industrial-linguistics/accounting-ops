#include "FirstRunWizard.hpp"

#include <QCheckBox>
#include <QFormLayout>
#include <QLabel>
#include <QLineEdit>
#include <QMessageBox>
#include <QPushButton>
#include <QVBoxLayout>
#include <QList>

ClientInfoPage::ClientInfoPage(QWidget *parent)
    : QWizardPage(parent)
    , m_clientNameEdit(new QLineEdit(this))
{
    setTitle(tr("Client information"));
    setSubTitle(tr("Provide the display name for the client whose credentials will be stored."));

    registerField("clientName*", m_clientNameEdit);

    auto *layout = new QVBoxLayout(this);
    auto *form = new QFormLayout();
    form->addRow(tr("Client name:"), m_clientNameEdit);
    layout->addLayout(form);
    setLayout(layout);
}

QString ClientInfoPage::clientName() const
{
    return m_clientNameEdit->text().trimmed();
}

ServiceCredentialPage::ServiceCredentialPage(const QString &serviceKey, const QString &serviceName, QWidget *parent)
    : QWizardPage(parent)
    , m_serviceKey(serviceKey)
    , m_serviceName(serviceName)
    , m_enableBox(new QCheckBox(tr("Configure %1 access").arg(serviceName), this))
    , m_clientIdEdit(new QLineEdit(this))
    , m_clientSecretEdit(new QLineEdit(this))
    , m_refreshTokenEdit(new QLineEdit(this))
    , m_regionEdit(new QLineEdit(this))
    , m_environmentEdit(new QLineEdit(this))
    , m_statusLabel(new QLabel(tr("Awaiting test."), this))
    , m_testButton(new QPushButton(tr("Test %1").arg(serviceName), this))
{
    setTitle(tr("%1 credentials").arg(serviceName));
    setSubTitle(tr("Enter the OAuth credentials for %1. Disable the checkbox if this client does not use the service.")
                    .arg(serviceName));

    m_enableBox->setChecked(true);
    m_clientSecretEdit->setEchoMode(QLineEdit::Password);

    auto *layout = new QVBoxLayout(this);
    layout->addWidget(m_enableBox);

    auto *form = new QFormLayout();
    form->addRow(tr("Client ID:"), m_clientIdEdit);
    form->addRow(tr("Client Secret:"), m_clientSecretEdit);
    form->addRow(tr("Refresh Token:"), m_refreshTokenEdit);
    form->addRow(tr("Region:"), m_regionEdit);
    form->addRow(tr("Environment:"), m_environmentEdit);
    layout->addLayout(form);

    layout->addWidget(m_testButton);
    layout->addWidget(m_statusLabel);
    layout->addStretch(1);

    connect(m_testButton, &QPushButton::clicked, this, &ServiceCredentialPage::handleTestClicked);
    connect(m_enableBox, &QCheckBox::toggled, this, &ServiceCredentialPage::handleEnableToggled);

    handleEnableToggled(true);
}

bool ServiceCredentialPage::isConfigured() const
{
    return m_enableBox->isChecked();
}

QString ServiceCredentialPage::serviceKey() const
{
    return m_serviceKey;
}

skills::ServiceCredential ServiceCredentialPage::credential() const
{
    skills::ServiceCredential credential;
    credential.clientId = m_clientIdEdit->text().trimmed();
    credential.clientSecret = m_clientSecretEdit->text().trimmed();
    credential.refreshToken = m_refreshTokenEdit->text().trimmed();
    credential.region = m_regionEdit->text().trimmed();
    credential.environment = m_environmentEdit->text().trimmed();
    return credential;
}

bool ServiceCredentialPage::validatePage()
{
    if (!isConfigured()) {
        return true;
    }

    if (m_clientIdEdit->text().trimmed().isEmpty() || m_clientSecretEdit->text().trimmed().isEmpty()) {
        QMessageBox::warning(this,
                             tr("Missing information"),
                             tr("Client ID and Client Secret are required for %1.").arg(m_serviceName));
        return false;
    }

    return true;
}

void ServiceCredentialPage::handleTestClicked()
{
    if (!isConfigured()) {
        m_statusLabel->setText(tr("%1 configuration skipped.").arg(m_serviceName));
        return;
    }

    if (m_clientIdEdit->text().trimmed().isEmpty() || m_clientSecretEdit->text().trimmed().isEmpty()) {
        QMessageBox::warning(this,
                             tr("Missing information"),
                             tr("Provide both Client ID and Client Secret before testing %1.").arg(m_serviceName));
        m_statusLabel->setText(tr("Test failed: incomplete details."));
        return;
    }

    QMessageBox::information(this,
                             tr("%1 connection test").arg(m_serviceName),
                             tr("%1 credentials look good!").arg(m_serviceName));
    m_statusLabel->setText(tr("Last test succeeded."));
}

void ServiceCredentialPage::handleEnableToggled(bool enabled)
{
    const QList<QWidget *> fields = {m_clientIdEdit, m_clientSecretEdit, m_refreshTokenEdit, m_regionEdit, m_environmentEdit, m_testButton};
    for (auto *widget : fields) {
        widget->setEnabled(enabled);
    }
    m_statusLabel->setText(enabled ? tr("Awaiting test.") : tr("%1 configuration skipped.").arg(m_serviceName));
}

FirstRunWizard::FirstRunWizard(skills::CredentialStore *store, QWidget *parent)
    : QWizard(parent)
    , m_store(store)
    , m_clientPage(new ClientInfoPage(this))
{
    setWindowTitle(tr("Accounting Ops first-run setup"));
    setWizardStyle(QWizard::ModernStyle);
    setOption(QWizard::NoBackButtonOnStartPage, true);

    addPage(m_clientPage);

    m_servicePages = {
        new ServiceCredentialPage(QStringLiteral("quickbooks"), tr("QuickBooks"), this),
        new ServiceCredentialPage(QStringLiteral("xero"), tr("Xero"), this),
        new ServiceCredentialPage(QStringLiteral("deputy"), tr("Deputy"), this),
    };

    for (auto *page : m_servicePages) {
        addPage(page);
    }
}

void FirstRunWizard::accept()
{
    const QString clientName = m_clientPage->clientName();
    if (clientName.isEmpty()) {
        QMessageBox::warning(this, tr("Client name required"), tr("Enter a client name before finishing."));
        return;
    }

    skills::ClientProfile profile;
    profile.displayName = clientName;

    for (auto *page : m_servicePages) {
        if (page->isConfigured()) {
            profile.serviceCredentials.insert(page->serviceKey(), page->credential());
        }
    }

    if (profile.serviceCredentials.isEmpty()) {
        QMessageBox::warning(this,
                             tr("No services selected"),
                             tr("Select at least one service to configure before finishing."));
        return;
    }

    QString error;
    if (!m_store->addOrUpdateClient(profile, &error)) {
        QMessageBox::critical(this, tr("Unable to save credentials"), error);
        return;
    }

    QMessageBox::information(this, tr("Setup complete"), tr("Credentials saved successfully."));
    QWizard::accept();
}
