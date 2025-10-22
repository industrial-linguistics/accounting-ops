cmake_minimum_required(VERSION 3.20)

# Read version from VERSION file
file(READ "${CMAKE_CURRENT_SOURCE_DIR}/VERSION" ACCOUNTING_OPS_VERSION)
string(STRIP "${ACCOUNTING_OPS_VERSION}" ACCOUNTING_OPS_VERSION)

project(AccountingOps VERSION ${ACCOUNTING_OPS_VERSION} LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# Generate version header
configure_file(
  "${CMAKE_CURRENT_SOURCE_DIR}/version.h.in"
  "${CMAKE_CURRENT_BINARY_DIR}/version.h"
  @ONLY
)

option(USE_QT6 "Prefer Qt6 over Qt5" ON)

if(USE_QT6)
  find_package(Qt6 COMPONENTS Widgets Sql REQUIRED)
  set(QT_VERSION_MAJOR 6)
else()
  find_package(Qt5 COMPONENTS Widgets Sql REQUIRED)
  set(QT_VERSION_MAJOR 5)
endif()

# Remove deprecated AGL framework from WrapOpenGL on macOS
# AGL was deprecated in macOS 10.14 and removed from modern SDKs
if(APPLE AND TARGET WrapOpenGL::WrapOpenGL)
  get_target_property(_opengl_libs WrapOpenGL::WrapOpenGL INTERFACE_LINK_LIBRARIES)
  if(_opengl_libs)
    list(FILTER _opengl_libs EXCLUDE REGEX ".*AGL.*")
    set_target_properties(WrapOpenGL::WrapOpenGL PROPERTIES INTERFACE_LINK_LIBRARIES "${_opengl_libs}")
  endif()
endif()


add_subdirectory(skills)
add_subdirectory(tools)

install(DIRECTORY docs/man/ DESTINATION share/man)
install(DIRECTORY docs/help/ DESTINATION share/accounting-ops/help)
install(DIRECTORY skills/data/ DESTINATION share/accounting-ops/skills)

set(CPACK_PACKAGE_NAME "accounting-ops")
set(CPACK_PACKAGE_VENDOR "Accounting Ops Toolkit")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "Cross-platform Qt tools and skills for accounting integrations")
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
set(CPACK_RESOURCE_FILE_LICENSE ${CMAKE_CURRENT_SOURCE_DIR}/LICENSE.txt)

if(WIN32)
  set(CPACK_GENERATOR "ZIP")
elseif(APPLE)
  set(CPACK_GENERATOR "productbuild")
  set(CPACK_PACKAGING_INSTALL_PREFIX "/usr/local")
  set(CPACK_MONOLITHIC_INSTALL 1)

  # Configure custom distribution file
  configure_file(${CMAKE_CURRENT_SOURCE_DIR}/distribution.dist.in
                 ${CMAKE_CURRENT_BINARY_DIR}/distribution.dist @ONLY)

  # Add custom target to rebuild package with proper distribution
  add_custom_target(package-macos
    COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} --target package
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_BINARY_DIR}/distribution.dist
            ${CMAKE_BINARY_DIR}/_CPack_Packages/Darwin/productbuild/accounting-ops-${PROJECT_VERSION}-Darwin/Contents/distribution.dist
    COMMAND productbuild --distribution ${CMAKE_BINARY_DIR}/_CPack_Packages/Darwin/productbuild/accounting-ops-${PROJECT_VERSION}-Darwin/Contents/distribution.dist
            --package-path ${CMAKE_BINARY_DIR}/_CPack_Packages/Darwin/productbuild/accounting-ops-${PROJECT_VERSION}-Darwin/Contents/Packages
            --resources ${CMAKE_BINARY_DIR}/_CPack_Packages/Darwin/productbuild/accounting-ops-${PROJECT_VERSION}-Darwin/Contents
            ${CMAKE_BINARY_DIR}/accounting-ops-${PROJECT_VERSION}-Darwin.pkg
    COMMENT "Building macOS package with proper distribution file"
    VERBATIM
  )
elseif(HAIKU)
  set(CPACK_GENERATOR "Package")
else()
  set(CPACK_GENERATOR "TGZ")
endif()

include(CPack)
