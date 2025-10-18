#include "SkillEditorWindow.hpp"

#include "skills/SkillRepository.hpp"

#include <QAction>
#include <QDir>
#include <QFileDialog>
#include <QFileInfo>
#include <QListWidget>
#include <QMessageBox>
#include <QPlainTextEdit>
#include <QPushButton>
#include <QStatusBar>
#include <QToolBar>
#include <QVBoxLayout>

SkillEditorWindow::SkillEditorWindow(skills::SkillRepository *repository, QWidget *parent)
    : QMainWindow(parent)
    , m_repository(repository)
    , m_skillList(new QListWidget(this))
    , m_editor(new QPlainTextEdit(this))
    , m_saveButton(new QPushButton(tr("Save"), this))
{
    setWindowTitle(tr("Skill File Editor"));
    m_editor->setPlaceholderText(tr("Select a skill file to begin editing."));

    auto *central = new QWidget(this);
    auto *layout = new QVBoxLayout(central);
    layout->addWidget(m_skillList);
    layout->addWidget(m_editor);
    layout->addWidget(m_saveButton);
    setCentralWidget(central);

    auto *openAction = new QAction(tr("Open skill directory"), this);
    connect(openAction, &QAction::triggered, this, [this]() {
        const auto path = QFileDialog::getExistingDirectory(this, tr("Select skill directory"));
        if (!path.isEmpty()) {
            QString error;
            if (!m_repository->loadFromDirectory(QDir(path), &error)) {
                QMessageBox::warning(this, tr("Unable to load skills"), error);
                return;
            }
            populateSkills();
            statusBar()->showMessage(tr("Loaded skills from %1").arg(path), 3000);
        }
    });

    auto *reloadAction = new QAction(tr("Reload"), this);
    connect(reloadAction, &QAction::triggered, this, &SkillEditorWindow::handleReload);

    auto *toolbar = addToolBar(tr("Skills"));
    toolbar->addAction(openAction);
    toolbar->addAction(reloadAction);

    connect(m_skillList, &QListWidget::currentTextChanged,
            this, &SkillEditorWindow::handleSkillSelection);
    connect(m_saveButton, &QPushButton::clicked,
            this, &SkillEditorWindow::handleSave);

    populateSkills();
}

void SkillEditorWindow::populateSkills() {
    m_skillList->clear();
    const auto docs = m_repository->skills();
    for (const auto &doc : docs) {
        m_skillList->addItem(doc.name);
    }
    if (!docs.isEmpty()) {
        m_skillList->setCurrentRow(0);
        loadSkill(docs.first());
    }
}

void SkillEditorWindow::loadSkill(const skills::SkillDocument &doc) {
    m_currentPath = doc.filePath;
    m_editor->setPlainText(doc.contents);
    statusBar()->showMessage(tr("Editing %1").arg(doc.name), 2000);
}

void SkillEditorWindow::handleSkillSelection(const QString &name) {
    if (name.isEmpty()) {
        m_currentPath.clear();
        m_editor->clear();
        return;
    }
    const auto docs = m_repository->skills();
    for (const auto &doc : docs) {
        if (doc.name == name) {
            loadSkill(doc);
            return;
        }
    }
}

void SkillEditorWindow::handleSave() {
    if (m_currentPath.isEmpty()) {
        QMessageBox::information(this, tr("No skill selected"), tr("Please select a skill to save."));
        return;
    }
    skills::SkillDocument doc;
    doc.filePath = m_currentPath;
    doc.name = QFileInfo(m_currentPath).completeBaseName();
    doc.contents = m_editor->toPlainText();
    QString error;
    if (!m_repository->saveSkill(doc, &error)) {
        QMessageBox::critical(this, tr("Unable to save"), error);
        return;
    }
    statusBar()->showMessage(tr("Saved %1").arg(doc.name), 2000);
}

void SkillEditorWindow::handleReload() {
    QString error;
    if (!m_repository->reload(&error)) {
        QMessageBox::warning(this, tr("Reload failed"), error);
        return;
    }
    populateSkills();
    statusBar()->showMessage(tr("Reloaded skill repository"), 2000);
}
