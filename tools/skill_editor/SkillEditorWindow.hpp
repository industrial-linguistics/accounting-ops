#pragma once

#include <QMainWindow>

namespace skills {
class SkillRepository;
struct SkillDocument;
}

class QListWidget;
class QPlainTextEdit;
class QPushButton;

class SkillEditorWindow : public QMainWindow {
    Q_OBJECT
public:
    SkillEditorWindow(skills::SkillRepository *repository, QWidget *parent = nullptr);

private slots:
    void handleSkillSelection(const QString &name);
    void handleSave();
    void handleReload();

private:
    void populateSkills();
    void loadSkill(const skills::SkillDocument &doc);

    skills::SkillRepository *m_repository;
    QListWidget *m_skillList;
    QPlainTextEdit *m_editor;
    QPushButton *m_saveButton;
    QString m_currentPath;
};

