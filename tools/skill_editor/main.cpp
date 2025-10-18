#include "SkillEditorWindow.hpp"

#include "skills/SkillRepository.hpp"

#include <QApplication>
#include <QCommandLineParser>
#include <QDir>
#include <QMessageBox>

int main(int argc, char *argv[]) {
    QApplication app(argc, argv);
    QCoreApplication::setApplicationName("SkillEditor");
    QCoreApplication::setApplicationVersion("1.0");

    QCommandLineParser parser;
    parser.setApplicationDescription("Qt-based editor for skill definition files");
    parser.addHelpOption();
    parser.addVersionOption();
    QCommandLineOption pathOption({"p", "path"},
                                  "Path to the skill directory.",
                                  "path");
    parser.addOption(pathOption);
    parser.process(app);

    skills::SkillRepository repository;
    QString error;
    const auto skillPath = parser.value(pathOption);
    if (!skillPath.isEmpty()) {
        if (!repository.loadFromDirectory(QDir(skillPath), &error)) {
            QMessageBox::warning(nullptr, QObject::tr("Unable to load skill repository"), error);
        }
    } else {
        const QDir defaultDir(QCoreApplication::applicationDirPath() + "/../skills/data");
        if (defaultDir.exists()) {
            repository.loadFromDirectory(defaultDir);
        }
    }

    SkillEditorWindow window(&repository);
    window.resize(720, 540);
    window.show();
    return app.exec();
}
