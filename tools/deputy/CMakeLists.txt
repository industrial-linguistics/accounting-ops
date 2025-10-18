set(DEPUTY_SOURCES
    main.cpp
    DeputyWindow.cpp
)

add_executable(deputy_tool ${DEPUTY_SOURCES})

target_include_directories(deputy_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(deputy_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS deputy_tool RUNTIME DESTINATION bin)
