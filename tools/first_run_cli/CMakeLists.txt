set(FIRST_RUN_CLI_SOURCES
    main.cpp
)

add_executable(first_run_cli_tool ${FIRST_RUN_CLI_SOURCES})

target_link_libraries(first_run_cli_tool
    PRIVATE
        skills_core
        Qt${QT_VERSION_MAJOR}::Core
)

install(TARGETS first_run_cli_tool RUNTIME DESTINATION bin)
