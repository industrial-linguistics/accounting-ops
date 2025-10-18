set(COMMON_SOURCES
    ConnectionTestWidget.cpp
    CredentialSelector.cpp
)

add_library(tooling_common STATIC ${COMMON_SOURCES})

target_include_directories(tooling_common
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(tooling_common
    PUBLIC
        skills_core
        Qt${QT_VERSION_MAJOR}::Widgets
)

install(TARGETS tooling_common
        ARCHIVE DESTINATION lib
        LIBRARY DESTINATION lib
        RUNTIME DESTINATION bin)

install(DIRECTORY include/ DESTINATION include)
