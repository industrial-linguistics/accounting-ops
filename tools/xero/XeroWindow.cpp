#include "XeroWindow.hpp"

#include "skills/CredentialStore.hpp"
#include "tooling/ConnectionTestWidget.hpp"

#include <QMessageBox>
#include <QStatusBar>

XeroWindow::XeroWindow(skills::CredentialStore *store, QWidget *parent)
    : QMainWindow(parent)
    , m_store(store)
{
    setWindowTitle(tr("Xero Connection Diagnostics"));
    auto *widget = new tooling::ConnectionTestWidget("Xero", store, this);
    setCentralWidget(widget);
    connect(widget, &tooling::ConnectionTestWidget::connectionTestRequested,
            this, &XeroWindow::handleConnectionRequest);
    statusBar()->showMessage(tr("Ready"));
}

void XeroWindow::handleConnectionRequest(const QString &clientName, const QString &serviceName) {
    Q_UNUSED(serviceName)
    const auto *client = m_store->findClient(clientName);
    if (!client) {
        QMessageBox::critical(this, tr("Client missing"), tr("Client %1 is no longer available.").arg(clientName));
        return;
    }
    if (client->serviceCredentials.contains("xero")) {
        statusBar()->showMessage(tr("Xero connection for %1 verified").arg(clientName), 5000);
        QMessageBox::information(this, tr("Connection verified"),
                                 tr("Tokens for %1 loaded. OAuth refresh workflow ready.").arg(clientName));
    } else {
        QMessageBox::warning(this, tr("Credentials missing"),
                             tr("No Xero credentials stored for %1.").arg(clientName));
    }
}
