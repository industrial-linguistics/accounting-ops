set(DEPUTY_SOURCES
    main.cpp
    DeputyWindow.cpp
)

if(APPLE)
    add_executable(deputy_tool MACOSX_BUNDLE ${DEPUTY_SOURCES})
    set_target_properties(deputy_tool PROPERTIES
        MACOSX_BUNDLE_BUNDLE_NAME "Deputy Tool"
        MACOSX_BUNDLE_GUI_IDENTIFIER "com.accountingops.deputy"
        MACOSX_BUNDLE_BUNDLE_VERSION "1.0.0"
        MACOSX_BUNDLE_SHORT_VERSION_STRING "1.0"
    )
else()
    add_executable(deputy_tool ${DEPUTY_SOURCES})
endif()

target_include_directories(deputy_tool PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(deputy_tool
    PRIVATE
        tooling_common
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

if(APPLE)
    set_target_properties(deputy_tool PROPERTIES
        INSTALL_RPATH "@executable_path/../Frameworks"
    )
    install(TARGETS deputy_tool BUNDLE DESTINATION "../../Applications")
else()
    install(TARGETS deputy_tool RUNTIME DESTINATION bin)
endif()
