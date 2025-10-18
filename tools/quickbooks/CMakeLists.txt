set(QUICKBOOKS_SOURCES
    main.cpp
    QuickBooksWindow.cpp
)

add_executable(quickbooks_tool ${QUICKBOOKS_SOURCES})

target_include_directories(quickbooks_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(quickbooks_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS quickbooks_tool RUNTIME DESTINATION bin)
