#include "tooling/CredentialSelector.hpp"

#include "skills/CredentialStore.hpp"

#include <QHBoxLayout>
#include <QBoxLayout>

namespace tooling {

CredentialSelector::CredentialSelector(skills::CredentialStore *store, QWidget *parent)
    : QWidget(parent)
    , m_store(store)
    , m_clientCombo(new QComboBox(this))
    , m_serviceCombo(new QComboBox(this))
{
    auto *layout = new QHBoxLayout(this);
    layout->addWidget(m_clientCombo);
    layout->addWidget(m_serviceCombo);

    connect(m_clientCombo, &QComboBox::currentTextChanged, this, [this](const QString &client) {
        m_serviceCombo->clear();
        const auto services = m_store->servicesForClient(client);
        m_serviceCombo->addItems(services);
        emit selectionChanged(client, m_serviceCombo->currentText());
    });

    connect(m_serviceCombo, &QComboBox::currentTextChanged, this, [this](const QString &service) {
        emit selectionChanged(m_clientCombo->currentText(), service);
    });

    rebuild();
}

QString CredentialSelector::selectedClient() const {
    return m_clientCombo->currentText();
}

QString CredentialSelector::selectedService() const {
    return m_serviceCombo->currentText();
}

void CredentialSelector::rebuild() {
    m_clientCombo->clear();
    m_serviceCombo->clear();
    const auto clients = m_store->clients();
    for (const auto &client : clients) {
        m_clientCombo->addItem(client.displayName);
    }
    if (!clients.isEmpty()) {
        const auto services = clients.first().serviceCredentials.keys();
        m_serviceCombo->addItems(services);
    }
}

} // namespace tooling
