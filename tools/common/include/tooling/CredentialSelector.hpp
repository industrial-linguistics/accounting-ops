#pragma once

#include <QComboBox>
#include <QWidget>

namespace skills {
class CredentialStore;
}

namespace tooling {

class CredentialSelector : public QWidget {
    Q_OBJECT
public:
    explicit CredentialSelector(skills::CredentialStore *store, QWidget *parent = nullptr);

    QString selectedClient() const;
    QString selectedService() const;

signals:
    void selectionChanged(const QString &client, const QString &service);

private:
    void rebuild();

    skills::CredentialStore *m_store;
    QComboBox *m_clientCombo;
    QComboBox *m_serviceCombo;
};

} // namespace tooling
