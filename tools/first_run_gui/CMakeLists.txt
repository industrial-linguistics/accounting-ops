set(FIRST_RUN_GUI_SOURCES
    FirstRunWizard.cpp
    main.cpp
)

if(APPLE)
    add_executable(first_run_gui_tool MACOSX_BUNDLE ${FIRST_RUN_GUI_SOURCES})
    set_target_properties(first_run_gui_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "First Run Setup"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.firstrun"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(first_run_gui_tool ${FIRST_RUN_GUI_SOURCES})
endif()

target_include_directories(first_run_gui_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(first_run_gui_tool
    PRIVATE
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(first_run_gui_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS first_run_gui_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS first_run_gui_tool RUNTIME DESTINATION bin)
endif()
