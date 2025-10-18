set(CLIENT_MANAGER_SOURCES
    main.cpp
    ClientManagerWindow.cpp
)

add_executable(client_manager_tool ${CLIENT_MANAGER_SOURCES})

target_include_directories(client_manager_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(client_manager_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS client_manager_tool RUNTIME DESTINATION bin)
