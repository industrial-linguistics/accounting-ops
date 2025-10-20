set(FIRST_RUN_GUI_SOURCES
    FirstRunWizard.cpp
    main.cpp
)

add_executable(first_run_gui_tool ${FIRST_RUN_GUI_SOURCES})

target_include_directories(first_run_gui_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(first_run_gui_tool
    PRIVATE
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS first_run_gui_tool RUNTIME DESTINATION bin)
