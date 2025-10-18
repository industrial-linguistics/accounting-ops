#include "ClientManagerWindow.hpp"

#include "skills/CredentialStore.hpp"

#include <QAction>
#include <QJsonDocument>
#include <QJsonObject>
#include <QListWidget>
#include <QMessageBox>
#include <QStatusBar>
#include <QTextEdit>
#include <QToolBar>
#include <QVBoxLayout>

ClientManagerWindow::ClientManagerWindow(skills::CredentialStore *store, QWidget *parent)
    : QMainWindow(parent)
    , m_store(store)
    , m_clientList(new QListWidget(this))
    , m_detailView(new QTextEdit(this))
{
    setWindowTitle(tr("Client Credential Manager"));
    m_detailView->setReadOnly(true);

    auto *central = new QWidget(this);
    auto *layout = new QVBoxLayout(central);
    layout->addWidget(m_clientList);
    layout->addWidget(m_detailView);
    setCentralWidget(central);

    auto *refreshAction = new QAction(tr("Refresh"), this);
    connect(refreshAction, &QAction::triggered, this, &ClientManagerWindow::handleRefresh);

    auto *toolbar = addToolBar(tr("Controls"));
    toolbar->addAction(refreshAction);

    connect(m_clientList, &QListWidget::currentTextChanged,
            this, &ClientManagerWindow::handleClientSelection);

    populateClients();
    statusBar()->showMessage(tr("Loaded %1 clients").arg(m_clientList->count()));
}

void ClientManagerWindow::populateClients() {
    m_clientList->clear();
    const auto clients = m_store->clients();
    for (const auto &client : clients) {
        m_clientList->addItem(client.displayName);
    }
    if (!clients.isEmpty()) {
        m_clientList->setCurrentRow(0);
        displayClient(clients.first());
    } else {
        m_detailView->setPlainText(tr("No clients configured."));
    }
}

void ClientManagerWindow::displayClient(const skills::ClientProfile &profile) {
    QJsonObject json;
    QJsonObject services;
    for (auto it = profile.serviceCredentials.cbegin(); it != profile.serviceCredentials.cend(); ++it) {
        QJsonObject service;
        service.insert("clientId", it.value().clientId);
        service.insert("environment", it.value().environment);
        service.insert("region", it.value().region);
        services.insert(it.key(), service);
    }
    json.insert("name", profile.displayName);
    json.insert("services", services);
    const auto text = QJsonDocument(json).toJson(QJsonDocument::Indented);
    m_detailView->setPlainText(QString::fromUtf8(text));
}

void ClientManagerWindow::handleClientSelection() {
    const auto name = m_clientList->currentItem() ? m_clientList->currentItem()->text() : QString();
    if (name.isEmpty()) {
        m_detailView->clear();
        return;
    }
    const auto *profile = m_store->findClient(name);
    if (!profile) {
        QMessageBox::warning(this, tr("Missing client"), tr("Client %1 could not be found.").arg(name));
        return;
    }
    displayClient(*profile);
}

void ClientManagerWindow::handleRefresh() {
    populateClients();
    statusBar()->showMessage(tr("Refreshed"), 2000);
}
