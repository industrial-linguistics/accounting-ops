#include "tooling/ConnectionTestWidget.hpp"

#include "skills/CredentialStore.hpp"

#include <QBoxLayout>
#include <QVBoxLayout>
#include <QMessageBox>

namespace tooling {

ConnectionTestWidget::ConnectionTestWidget(const QString &serviceName, skills::CredentialStore *store, QWidget *parent)
    : QWidget(parent)
    , m_serviceName(serviceName)
    , m_store(store)
    , m_clientInput(new QLineEdit(this))
    , m_statusLabel(new QLabel(tr("Waiting for test"), this))
{
    auto *layout = new QVBoxLayout(this);
    auto *instructions = new QLabel(tr("Enter the client name to test the %1 connection.").arg(serviceName), this);
    auto *testButton = new QPushButton(tr("Test Connection"), this);

    layout->addWidget(instructions);
    layout->addWidget(m_clientInput);
    layout->addWidget(testButton);
    layout->addWidget(m_statusLabel);

    connect(testButton, &QPushButton::clicked, this, &ConnectionTestWidget::triggerTest);
}

void ConnectionTestWidget::triggerTest() {
    const QString clientName = m_clientInput->text().trimmed();
    if (clientName.isEmpty()) {
        QMessageBox::warning(this, tr("Missing client"), tr("Please enter a client name."));
        return;
    }

    if (!m_store->findClient(clientName)) {
        QMessageBox::warning(this, tr("Unknown client"), tr("No credentials found for %1.").arg(clientName));
        return;
    }

    emit connectionTestRequested(clientName, m_serviceName);
    m_statusLabel->setText(tr("Test requested for %1").arg(clientName));
}

} // namespace tooling
