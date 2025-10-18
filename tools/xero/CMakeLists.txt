set(XERO_SOURCES
    main.cpp
    XeroWindow.cpp
)

add_executable(xero_tool ${XERO_SOURCES})

target_include_directories(xero_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(xero_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS xero_tool RUNTIME DESTINATION bin)
