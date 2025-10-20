#include "ClientManagerWindow.hpp"

#include "skills/CredentialStore.hpp"

#include <QApplication>
#include <QCommandLineParser>
#include <QDir>
#include <QMessageBox>

static QString resolveDefaultCredentialPath() {
    QDir dir(QCoreApplication::applicationDirPath());
    dir.cdUp();
    return dir.filePath("config/credentials.sqlite");
}

int main(int argc, char *argv[]) {
    QApplication app(argc, argv);
    QCoreApplication::setApplicationName("ClientManager");
    QCoreApplication::setApplicationVersion("1.0");

    QCommandLineParser parser;
    parser.setApplicationDescription("Manage multi-client credential sets for Deputy, Xero, and QuickBooks");
    parser.addHelpOption();
    parser.addVersionOption();
    QCommandLineOption credentialsOption({"c", "credentials"},
                                         "Path to the shared credentials database file.",
                                         "file");
    parser.addOption(credentialsOption);
    parser.process(app);

    QString credentialPath = parser.value(credentialsOption);
    if (credentialPath.isEmpty()) {
        credentialPath = resolveDefaultCredentialPath();
    }

    skills::CredentialStore store;
    QString error;
    if (!credentialPath.isEmpty() && !store.loadFromFile(credentialPath, &error)) {
        QMessageBox::warning(nullptr, QObject::tr("Credentials not loaded"), error);
    }

    ClientManagerWindow window(&store);
    window.resize(640, 480);
    window.show();
    return app.exec();
}
