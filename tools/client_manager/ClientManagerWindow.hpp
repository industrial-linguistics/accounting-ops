#pragma once

#include <QMainWindow>

namespace skills {
class CredentialStore;
struct ClientProfile;
}

class QListWidget;
class QTextEdit;

class ClientManagerWindow : public QMainWindow {
    Q_OBJECT
public:
    explicit ClientManagerWindow(skills::CredentialStore *store, QWidget *parent = nullptr);

private slots:
    void handleClientSelection();
    void handleRefresh();

private:
    void populateClients();
    void displayClient(const skills::ClientProfile &profile);

    skills::CredentialStore *m_store;
    QListWidget *m_clientList;
    QTextEdit *m_detailView;
};

