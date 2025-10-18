#include "XeroWindow.hpp"

#include "skills/CredentialStore.hpp"

#include <QApplication>
#include <QCommandLineParser>
#include <QDir>
#include <QMessageBox>

static QString resolveDefaultCredentialPath() {
    QDir dir(QCoreApplication::applicationDirPath());
    dir.cdUp();
    if (dir.exists("config/credentials.json")) {
        return dir.filePath("config/credentials.json");
    }
    return QString();
}

int main(int argc, char *argv[]) {
    QApplication app(argc, argv);
    QCoreApplication::setApplicationName("XeroTool");
    QCoreApplication::setApplicationVersion("1.0");

    QCommandLineParser parser;
    parser.setApplicationDescription("Xero connection diagnostic tool");
    parser.addHelpOption();
    parser.addVersionOption();
    QCommandLineOption credentialsOption({"c", "credentials"},
                                         "Path to the shared credentials JSON file.",
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

    XeroWindow window(&store);
    window.resize(480, 240);
    window.show();
    return app.exec();
}
