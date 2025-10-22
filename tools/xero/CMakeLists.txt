set(XERO_SOURCES
    main.cpp
    XeroWindow.cpp
)

if(APPLE)
    add_executable(xero_tool MACOSX_BUNDLE ${XERO_SOURCES})
    set_target_properties(xero_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "Xero Tool"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.xero"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(xero_tool ${XERO_SOURCES})
endif()

target_include_directories(xero_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR} ${CMAKE_BINARY_DIR})

target_link_libraries(xero_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(xero_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS xero_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS xero_tool RUNTIME DESTINATION bin)
endif()
