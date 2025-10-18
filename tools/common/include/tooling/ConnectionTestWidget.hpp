#pragma once

#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QWidget>

namespace skills {
class CredentialStore;
}

namespace tooling {

class ConnectionTestWidget : public QWidget {
    Q_OBJECT
public:
    explicit ConnectionTestWidget(const QString &serviceName, skills::CredentialStore *store, QWidget *parent = nullptr);

signals:
    void connectionTestRequested(const QString &clientName, const QString &serviceName);

private slots:
    void triggerTest();

private:
    QString m_serviceName;
    skills::CredentialStore *m_store;
    QLineEdit *m_clientInput;
    QLabel *m_statusLabel;
};

} // namespace tooling
