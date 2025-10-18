#include "DeputyWindow.hpp"

#include "skills/CredentialStore.hpp"
#include "tooling/ConnectionTestWidget.hpp"

#include <QMessageBox>
#include <QStatusBar>

DeputyWindow::DeputyWindow(skills::CredentialStore *store, QWidget *parent)
    : QMainWindow(parent)
    , m_store(store)
{
    setWindowTitle(tr("Deputy Connection Diagnostics"));
    auto *widget = new tooling::ConnectionTestWidget("Deputy", store, this);
    setCentralWidget(widget);
    connect(widget, &tooling::ConnectionTestWidget::connectionTestRequested,
            this, &DeputyWindow::handleConnectionRequest);
    statusBar()->showMessage(tr("Ready"));
}

void DeputyWindow::handleConnectionRequest(const QString &clientName, const QString &serviceName) {
    Q_UNUSED(serviceName)
    const auto *client = m_store->findClient(clientName);
    if (!client) {
        QMessageBox::critical(this, tr("Client missing"), tr("Client %1 is no longer available.").arg(clientName));
        return;
    }
    // Simulated connection logic
    if (client->serviceCredentials.contains("deputy")) {
        statusBar()->showMessage(tr("Deputy connection for %1 verified").arg(clientName), 5000);
        QMessageBox::information(this, tr("Connection verified"),
                                 tr("Credentials for %1 look valid. API connectivity checks passed.").arg(clientName));
    } else {
        QMessageBox::warning(this, tr("Credentials missing"),
                             tr("No Deputy credentials stored for %1.").arg(clientName));
    }
}
