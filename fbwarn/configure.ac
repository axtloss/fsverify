AC_INIT([fbwarn], [1.0], [])
AM_INIT_AUTOMAKE([-Wall -Werror foreign])
AC_PROG_CC
AC_CHECK_FUNCS([
    strdup
    strstr
])
AC_CHECK_HEADERS([
    unistd.h
    raylib.h
])
AC_CHECK_HEADER_STDBOOL
AC_FUNC_MALLOC
AC_FUNC_REALLOC
AC_TYPE_SIZE_T
AC_TYPE_SSIZE_T
AC_CONFIG_FILES([
    Makefile
    src/Makefile
])
AC_OUTPUT
