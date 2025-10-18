#pragma once

#include <QMainWindow>

namespace skills {
class CredentialStore;
}

namespace tooling {
class ConnectionTestWidget;
}

class XeroWindow : public QMainWindow {
    Q_OBJECT
public:
    explicit XeroWindow(skills::CredentialStore *store, QWidget *parent = nullptr);

private slots:
    void handleConnectionRequest(const QString &clientName, const QString &serviceName);

private:
    skills::CredentialStore *m_store;
};

