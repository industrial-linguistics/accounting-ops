set(SKILL_EDITOR_SOURCES
    main.cpp
    SkillEditorWindow.cpp
)

add_executable(skill_editor_tool ${SKILL_EDITOR_SOURCES})

target_include_directories(skill_editor_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(skill_editor_tool
    PRIVATE
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS skill_editor_tool RUNTIME DESTINATION bin)
