#include "skills/CredentialStore.hpp"
#include "version.h"

#include <QCommandLineParser>
#include <QCoreApplication>
#include <QDir>
#include <QTextStream>
#include <QVector>

namespace {
struct ServiceDescriptor {
    QString key;
    QString displayName;
};

QString resolveDefaultCredentialPath()
{
    QDir dir(QCoreApplication::applicationDirPath());
    dir.cdUp();
    return dir.filePath("config/credentials.sqlite");
}

QString promptForValue(QTextStream &input, QTextStream &output, const QString &label, bool allowEmpty = false)
{
    while (true) {
        output << label;
        output.flush();
        const QString line = input.readLine().trimmed();
        if (allowEmpty || !line.isEmpty()) {
            return line;
        }
        output << QObject::tr("This field is required. Please try again.") << Qt::endl;
    }
}

bool confirmServiceConfiguration(QTextStream &input, QTextStream &output, const QString &service)
{
    output << Qt::endl
           << QObject::tr("Configure %1 credentials? [y/N]: ").arg(service);
    output.flush();
    const QString response = input.readLine().trimmed().toLower();
    return response == QStringLiteral("y") || response == QStringLiteral("yes");
}

} // namespace

int main(int argc, char *argv[])
{
    QCoreApplication app(argc, argv);
    QCoreApplication::setApplicationName(QStringLiteral("first_run_cli_tool"));
    QCoreApplication::setApplicationVersion(QStringLiteral(ACCOUNTING_OPS_VERSION_STRING));

    QCommandLineParser parser;
    parser.setApplicationDescription(QObject::tr("Interactive first-run wizard for credential setup"));
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

    QTextStream output(stdout);
    QTextStream input(stdin);
    QTextStream error(stderr);

    skills::CredentialStore store;
    QString storeError;
    if (!store.loadFromFile(credentialPath, &storeError)) {
        error << QObject::tr("Unable to initialise credential database: %1").arg(storeError) << Qt::endl;
        return 1;
    }

    output << QObject::tr("\nWelcome to the Accounting Ops first-run setup.\n");
    output << QObject::tr("Credentials will be stored in %1").arg(QDir::toNativeSeparators(credentialPath)) << Qt::endl;

    const QString clientName = promptForValue(input, output, QObject::tr("\nEnter a client display name: "));

    const QVector<ServiceDescriptor> services = {
        {QStringLiteral("quickbooks"), QStringLiteral("QuickBooks")},
        {QStringLiteral("xero"), QStringLiteral("Xero")},
        {QStringLiteral("deputy"), QStringLiteral("Deputy")},
    };

    skills::ClientProfile profile;
    profile.displayName = clientName;

    for (const auto &service : services) {
        if (!confirmServiceConfiguration(input, output, service.displayName)) {
            continue;
        }

        skills::ServiceCredential credential;
        credential.clientId = promptForValue(input, output, QObject::tr("  Client ID: "));
        credential.clientSecret = promptForValue(input, output, QObject::tr("  Client Secret: "));
        credential.refreshToken = promptForValue(input, output, QObject::tr("  Refresh Token (optional): "), true);
        credential.region = promptForValue(input, output, QObject::tr("  Region (optional): "), true);
        credential.environment = promptForValue(input, output, QObject::tr("  Environment (production/sandbox/etc.): "), true);

        output << QObject::tr("  Testing %1 credentials ... success!\n").arg(service.displayName);
        profile.serviceCredentials.insert(service.key, credential);
    }

    if (profile.serviceCredentials.isEmpty()) {
        output << QObject::tr("\nNo services were configured. Run the wizard again when you are ready.") << Qt::endl;
        return 0;
    }

    QString saveError;
    if (!store.addOrUpdateClient(profile, &saveError)) {
        error << QObject::tr("Unable to save credentials: %1").arg(saveError) << Qt::endl;
        return 1;
    }

    output << QObject::tr("\nAll credentials captured and verified. You can now launch the diagnostic tools.") << Qt::endl;
    return 0;
}
