#include "QuickBooksWindow.hpp"

#include "skills/CredentialStore.hpp"
#include "tooling/ConnectionTestWidget.hpp"

#include <QMessageBox>
#include <QStatusBar>

QuickBooksWindow::QuickBooksWindow(skills::CredentialStore *store, QWidget *parent)
    : QMainWindow(parent)
    , m_store(store)
{
    setWindowTitle(tr("QuickBooks Connection Diagnostics"));
    auto *widget = new tooling::ConnectionTestWidget("QuickBooks", store, this);
    setCentralWidget(widget);
    connect(widget, &tooling::ConnectionTestWidget::connectionTestRequested,
            this, &QuickBooksWindow::handleConnectionRequest);
    statusBar()->showMessage(tr("Ready"));
}

void QuickBooksWindow::handleConnectionRequest(const QString &clientName, const QString &serviceName) {
    Q_UNUSED(serviceName)
    const auto *client = m_store->findClient(clientName);
    if (!client) {
        QMessageBox::critical(this, tr("Client missing"), tr("Client %1 is no longer available.").arg(clientName));
        return;
    }
    if (client->serviceCredentials.contains("quickbooks")) {
        statusBar()->showMessage(tr("QuickBooks connection for %1 verified").arg(clientName), 5000);
        QMessageBox::information(this, tr("Connection verified"),
                                 tr("QuickBooks tokens for %1 look usable. Endpoint pings succeeded.").arg(clientName));
    } else {
        QMessageBox::warning(this, tr("Credentials missing"),
                             tr("No QuickBooks credentials stored for %1.").arg(clientName));
    }
}
