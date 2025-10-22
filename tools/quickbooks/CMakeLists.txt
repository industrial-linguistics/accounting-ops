set(QUICKBOOKS_SOURCES
    main.cpp
    QuickBooksWindow.cpp
)

if(APPLE)
    add_executable(quickbooks_tool MACOSX_BUNDLE ${QUICKBOOKS_SOURCES})
    set_target_properties(quickbooks_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "QuickBooks Tool"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.quickbooks"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(quickbooks_tool ${QUICKBOOKS_SOURCES})
endif()

target_include_directories(quickbooks_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(quickbooks_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(quickbooks_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS quickbooks_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS quickbooks_tool RUNTIME DESTINATION bin)
endif()
