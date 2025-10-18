#pragma once

#include <QDir>
#include <QObject>
#include <QVector>

namespace skills {

struct SkillDocument {
    QString name;
    QString filePath;
    QString description;
    QString contents;
};

class SkillRepository : public QObject {
    Q_OBJECT
public:
    explicit SkillRepository(QObject *parent = nullptr);

    void clear();
    bool loadFromDirectory(const QDir &directory, QString *error = nullptr);
    bool reload(QString *error = nullptr);

    QVector<SkillDocument> skills() const;
    bool saveSkill(const SkillDocument &doc, QString *error = nullptr);
    bool removeSkill(const QString &filePath, QString *error = nullptr);

    QDir root() const;

signals:
    void repositoryChanged();

private:
    QDir m_root;
    QVector<SkillDocument> m_skills;
};

} // namespace skills
