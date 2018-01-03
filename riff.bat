@echo off
:: ##########################################################################
::  riff command line interface for Windows
:: ##########################################################################

set RIFF_VERSION=0.0.2

:: set local scope for the variables with windows NT shell
if [%OS%]==[Windows_NT] setlocal

:: handle general help
if [%1]==[] set _COMMAND=help
if [%1]==[--help] set _COMMAND=help
if [%1]==[-h] set _COMMAND=help
if [%_COMMAND%]==[help] (
    call :print_usage
    exit /B %ERRORLEVEL%
)

:: grab the command
set _COMMAND=%1
shift

:: handle commands and arguments/options
if [%_COMMAND%]==[build] goto arg_loop
if [%_COMMAND%]==[create] goto arg_loop
if [%_COMMAND%]==[apply] goto arg_loop
if [%_COMMAND%]==[delete] goto arg_loop
if [%_COMMAND%]==[list] goto arg_loop
if [%_COMMAND%]==[logs] goto arg_loop
if [%_COMMAND%]==[version] goto arg_loop

echo %_COMMAND% is an invalid command
echo.
echo Type "riff --help" to see valid commands
echo.
exit /B 1

:: parse arguments/options
:arg_loop
set key=%1
if [%key%]==[] goto end_args
set match=false
if [%1]==[--name] set key=-n
if [%1]==[--filename] set key=-f
if [%1]==[--version] set key=-v
if [%1]==[--container] set key=-c
if [%1]==[--tail] set key=-t
if [%1]==[--help] set key=-h
if [%key%]==[-n] (
    set match=true
    set FUNCTION=%2
    shift
)
if [%key%]==[-f] (
    set match=true
    set FNPATH=%2
    shift
)
if [%key%]==[-v] (
    set match=true
    set VERSION=%2
    shift
)
if [%key%]==[-c] (
    set match=true
    set CONTAINER=%2
    shift
)
if [%key%]==[-t] (
    set match=true
    set TAIL=-f
)
if [%key%]==[-h] (
    set match=true
    call :print_%_COMMAND%_usage
    exit /B 1
)
if [%match%]==[false] (
    echo.
    echo ERROR: Invalid option: [%key%]
    call :print_%_COMMAND%_usage
    exit /B 1
)
shift
goto arg_loop
:end_args

:: defaults
if [%FUNCTION%]==[] (
  for %%a in (.) do set FUNCTION=%%~na
)
if [%FNPATH%]==[] set FNPATH=.
if [%VERSION%]==[] set VERSION=0.0.1
if [%CONTAINER%]==[] set CONTAINER=sidecar

:: execute the commands
if [%_COMMAND%]==[build] (
    call :build
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[create] (
    call :create
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[apply] (
    call :apply
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[delete] (
    call :delete
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[list] (
    call :list
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[logs] (
    call :logs
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[version] (
    call :version
    exit /B %ERRORLEVEL%
)

exit /B 1

:: commands
:build
  echo Building %FUNCTION% %FNPATH% %VERSION%
  docker build -t projectriff/%FUNCTION%:%VERSION% %FNPATH%
  exit /B %ERRORLEVEL%

:create
  echo Creating %FNPATH% resource[s]
  kubectl create -f %FNPATH%
  exit /B %ERRORLEVEL%

:apply
  echo Applying %FNPATH% resource[s]
  kubectl apply -f %FNPATH%
  exit /B %ERRORLEVEL%

:delete
  echo Deleting %FUNCTION% resource[s]
  kubectl delete function %FUNCTION%
  exit /B %ERRORLEVEL%

:list
  echo Listing function resources
  kubectl get functions
  exit /B %ERRORLEVEL%

:logs
  echo Displaying logs for container %CONTAINER% of function %FUNCTION%
  set _tmp_file=%tmp%\riff-%random%.txt
  set _kcmd=kubectl get pod -l function=%FUNCTION% -o jsonpath="{.items[0].metadata.name}"
  %_kcmd% > %_tmp_file%
  set /p _pod= < %_tmp_file%
  del %_tmp_file%
  echo.
  kubectl logs %TAIL% -c %CONTAINER% %_pod%
  exit /B %ERRORLEVEL%

:version
  echo riff version %RIFF_VERSION%
  exit /B 0


:: help texts
:print_usage
echo.
echo riff is for functions
echo.
echo version %RIFF_VERSION%
echo.
echo Commands:
echo   build        Build a function container
echo   create       Create function resource definitions
echo   apply        Apply function resource definitions
echo   delete       Delete function resource definitions
echo   list         List current function resources
echo   logs         Show logs for a function resource
echo   version      Display the riff version
echo.
echo   Use "riff <command> --help" for more information about a given command.
echo.
exit /B 0

:print_build_usage
echo.
echo Build the function based on the code available in the path directory, using the name
echo   and version specified for the image that is built.
echo.
echo Usage:
echo.
echo   riff build -n ^<name^> -v ^<version^> -f ^<path^>
echo.
echo Options:
echo.
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -v, --version: the version of the function (defaults to 0.0.1)
echo   -f, --filename: filename, directory, or URL for the code or resource (defaults to the current directory)
echo.
exit /B 0

:print_create_usage
echo.
echo Create the resource[s] based on the resource definition[s] that the path points to.
echo.
echo Usage:
echo.
echo   riff create -f ^<path^>
echo.
echo Options:
echo.
echo   -f, --filename: filename, directory, or URL for the resource definition[s] (defaults to the current directory)
echo.
exit /B 0

:print_apply_usage
echo.
echo Apply the resource definition[s] that the path points to. A resource will be created if
echo   it doesn't exist yet.
echo.
echo Usage:
echo.
echo   riff apply -f ^<path^>
echo.
echo Options:
echo.
echo   -f, --filename: filename, directory, or URL for the resource definition[s] (defaults to the current directory)
echo.
exit /B 0

:print_delete_usage
echo.
echo Delete the resource definition[s] for the function specified.
echo.
echo Usage:
echo.
echo   riff delete -n ^<name^>
echo.
echo Options:
echo.
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo.
exit /B 0

:print_list_usage
echo.
echo List the current function resources.
echo.
echo Usage:
echo.
echo   riff list
echo.
exit /B 0

:print_logs_usage
echo.
echo Display the logs for a running function.
echo.
echo Usage:
echo.
echo   riff logs -n ^<name^> [-c ^<container^>] [-t]
echo.
echo Options:
echo.
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -c, --container: the name of the container, usually sidecar or main (defaults to sidecar)
echo   -t, --tail: tail the logs
echo.
exit /B 0

:print_version_usage
echo.
echo Display the riff version.
echo.
echo Usage:
echo.
echo   riff version
echo.
exit /B 0
