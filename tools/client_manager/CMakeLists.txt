set(CLIENT_MANAGER_SOURCES
    main.cpp
    ClientManagerWindow.cpp
)

if(APPLE)
    add_executable(client_manager_tool MACOSX_BUNDLE ${CLIENT_MANAGER_SOURCES})
    set_target_properties(client_manager_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "Client Manager"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.clientmanager"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(client_manager_tool ${CLIENT_MANAGER_SOURCES})
endif()

target_include_directories(client_manager_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR} ${CMAKE_BINARY_DIR})

target_link_libraries(client_manager_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(client_manager_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS client_manager_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS client_manager_tool RUNTIME DESTINATION bin)
endif()
