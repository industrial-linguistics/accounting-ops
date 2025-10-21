set(SKILLS_SOURCES
    src/CredentialStore.cpp
    src/SkillRepository.cpp
)

set(SKILLS_HEADERS
    include/skills/CredentialStore.hpp
    include/skills/SkillRepository.hpp
)

add_library(skills_core STATIC ${SKILLS_SOURCES} ${SKILLS_HEADERS})

target_include_directories(skills_core
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(skills_core
    PUBLIC
        Qt${QT_VERSION_MAJOR}::Core
        Qt${QT_VERSION_MAJOR}::Sql
)

install(TARGETS skills_core
        EXPORT SkillsTargets
        ARCHIVE DESTINATION lib
        LIBRARY DESTINATION lib
        RUNTIME DESTINATION bin)

install(DIRECTORY include/ DESTINATION include)

install(EXPORT SkillsTargets
        FILE SkillsTargets.cmake
        NAMESPACE Skills::
        DESTINATION lib/cmake/Skills)
