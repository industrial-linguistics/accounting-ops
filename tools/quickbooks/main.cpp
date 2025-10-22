#include "QuickBooksWindow.hpp"

#include "skills/CredentialStore.hpp"
#include "version.h"

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
    QCoreApplication::setApplicationName("QuickBooksTool");
    QCoreApplication::setApplicationVersion(ACCOUNTING_OPS_VERSION_STRING);

    QCommandLineParser parser;
    parser.setApplicationDescription("QuickBooks connection diagnostic tool");
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

    QuickBooksWindow window(&store);
    window.resize(480, 240);
    window.show();
    return app.exec();
}
