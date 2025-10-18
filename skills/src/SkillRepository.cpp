#include "skills/SkillRepository.hpp"

#include <QFile>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTextStream>

namespace skills {

SkillRepository::SkillRepository(QObject *parent)
    : QObject(parent) {}

void SkillRepository::clear() {
    if (!m_skills.isEmpty()) {
        m_skills.clear();
        emit repositoryChanged();
    }
    m_root = QDir();
}

bool SkillRepository::loadFromDirectory(const QDir &directory, QString *error) {
    if (!directory.exists()) {
        if (error) {
            *error = tr("Skill directory does not exist: %1").arg(directory.absolutePath());
        }
        return false;
    }
    m_root = directory;
    return reload(error);
}

bool SkillRepository::reload(QString *error) {
    if (!m_root.exists()) {
        if (error) {
            *error = tr("Skill directory has not been configured");
        }
        return false;
    }

    QVector<SkillDocument> loaded;

    const auto entries = m_root.entryInfoList({"*.skill.json", "*.skill"}, QDir::Files);
    for (const QFileInfo &info : entries) {
        QFile file(info.absoluteFilePath());
        if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
            if (error) {
                *error = tr("Unable to read skill file %1: %2").arg(info.fileName(), file.errorString());
            }
            return false;
        }
        QTextStream stream(&file);
        SkillDocument doc;
        doc.filePath = info.absoluteFilePath();
        doc.name = info.completeBaseName();
        doc.contents = stream.readAll();

        const auto jsonDoc = QJsonDocument::fromJson(doc.contents.toUtf8());
        if (jsonDoc.isObject()) {
            doc.description = jsonDoc.object().value("description").toString();
        }

        loaded.push_back(doc);
    }

    m_skills = loaded;
    emit repositoryChanged();
    return true;
}

QVector<SkillDocument> SkillRepository::skills() const {
    return m_skills;
}

bool SkillRepository::saveSkill(const SkillDocument &doc, QString *error) {
    if (doc.filePath.isEmpty()) {
        if (error) {
            *error = tr("Skill file path is empty");
        }
        return false;
    }

    QFile file(doc.filePath);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text | QIODevice::Truncate)) {
        if (error) {
            *error = tr("Unable to write skill file %1: %2").arg(doc.filePath, file.errorString());
        }
        return false;
    }

    QTextStream stream(&file);
    stream << doc.contents;
    file.close();

    for (auto &existing : m_skills) {
        if (existing.filePath == doc.filePath) {
            existing = doc;
            emit repositoryChanged();
            return true;
        }
    }

    m_skills.push_back(doc);
    emit repositoryChanged();
    return true;
}

bool SkillRepository::removeSkill(const QString &filePath, QString *error) {
    QFile file(filePath);
    if (file.exists() && !file.remove()) {
        if (error) {
            *error = tr("Unable to remove skill file %1: %2").arg(filePath, file.errorString());
        }
        return false;
    }

    for (int i = 0; i < m_skills.size(); ++i) {
        if (m_skills[i].filePath == filePath) {
            m_skills.removeAt(i);
            emit repositoryChanged();
            break;
        }
    }
    return true;
}

QDir SkillRepository::root() const {
    return m_root;
}

} // namespace skills
