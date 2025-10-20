#include "FirstRunWizard.hpp"

#include "skills/CredentialStore.hpp"

#include <QApplication>
#include <QCommandLineParser>
#include <QDir>
#include <QMessageBox>

static QString resolveDefaultCredentialPath()
{
    QDir dir(QCoreApplication::applicationDirPath());
    dir.cdUp();
    return dir.filePath("config/credentials.sqlite");
}

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    QCoreApplication::setApplicationName(QStringLiteral("first_run_gui_tool"));
    QCoreApplication::setApplicationVersion(QStringLiteral("1.0"));

    QCommandLineParser parser;
    parser.setApplicationDescription(QObject::tr("Guided first-run wizard for credential setup"));
    parser.addHelpOption();
    parser.addVersionOption();

    QCommandLineOption credentialsOption({QStringLiteral("c"), QStringLiteral("credentials")},
                                         QObject::tr("Path to the shared credentials database file."),
                                         QObject::tr("file"));
    parser.addOption(credentialsOption);
    parser.process(app);

    QString credentialPath = parser.value(credentialsOption);
    if (credentialPath.isEmpty()) {
        credentialPath = resolveDefaultCredentialPath();
    }

    skills::CredentialStore store;
    QString error;
    if (!store.loadFromFile(credentialPath, &error)) {
        QMessageBox::critical(nullptr,
                              QObject::tr("Unable to initialise credential database"),
                              error);
        return 1;
    }

    FirstRunWizard wizard(&store);
    wizard.resize(560, 460);
    wizard.show();
    return app.exec();
}
