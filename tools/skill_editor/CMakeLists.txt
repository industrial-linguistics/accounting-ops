set(SKILL_EDITOR_SOURCES
    main.cpp
    SkillEditorWindow.cpp
)

if(APPLE)
    add_executable(skill_editor_tool MACOSX_BUNDLE ${SKILL_EDITOR_SOURCES})
    set_target_properties(skill_editor_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "Skill Editor"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.skilleditor"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(skill_editor_tool ${SKILL_EDITOR_SOURCES})
endif()

target_include_directories(skill_editor_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR} ${CMAKE_BINARY_DIR})

target_link_libraries(skill_editor_tool
    PRIVATE
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(skill_editor_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS skill_editor_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS skill_editor_tool RUNTIME DESTINATION bin)
endif()
