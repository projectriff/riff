@echo off
:: ##########################################################################
::  riff command line interface for Windows
:: ##########################################################################

:: set local scope for the variables with windows NT shell
setlocal enabledelayedexpansion

if [%RIFF_VERSION%]==[] (
    set RIFF_VERSION=0.0.2
)

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
if [%_COMMAND%]==[init] goto arg_loop
if [%_COMMAND%]==[build] goto arg_loop
if [%_COMMAND%]==[create] goto arg_loop
if [%_COMMAND%]==[apply] goto arg_loop
if [%_COMMAND%]==[update] goto arg_loop
if [%_COMMAND%]==[delete] goto arg_loop
if [%_COMMAND%]==[list] goto arg_loop
if [%_COMMAND%]==[logs] goto arg_loop
if [%_COMMAND%]==[publish] goto arg_loop
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
if [%1]==[--language] set key=-l
if [%1]==[--protocol] set key=-p
if [%1]==[--artifact] set key=-a
if [%1]==[--container] set key=-c
if [%1]==[--input] set key=-i
if [%1]==[--output] set key=-o
if [%1]==[--useraccount] set key=-u
if [%1]==[--data] set key=-d
if [%1]==[--reply] set key=-r
if [%1]==[--eval] set key=-e
if [%1]==[--tail] set key=-t
if [%1]==[--help] set key=-h
if [%1]==[--riff-version] (
    set match=true
    set RIFF_VERSION=%2
    shift
)
if [%1]==[--handler] (
    set match=true
    set FNHANDLER=%2
    shift
)
if [%1]==[--push] (
    set match=true
    set DOCKERPUSH=true
)
if [%1]==[--all] (
    set match=true
    set DELETEALL=true
)
if [%1]==[--count] (
    set match=true
    set PUB_COUNT=%2
    shift
)
if [%1]==[--pause] (
    set match=true
    set PUB_PAUSE=%2
    shift
)
if [%key%]==[-n] (
    set match=true
    set FUNCTION=%2
    shift
)
if [%key%]==[-f] (
    set match=true
    set FNPATH=%~2
    shift
)
if [%key%]==[-v] (
    set match=true
    set VERSION=%2
    shift
)
if [%key%]==[-l] (
    set match=true
    set FNLANG=%2
    shift
)
if [%key%]==[-p] (
    set match=true
    set FNPROT=%2
    shift
)
if [%key%]==[-a] (
    set match=true
    set FNARTIFACT=%2
    shift
)
if [%key%]==[-c] (
    set match=true
    set CONTAINER=%2
    shift
)
if [%key%]==[-i] (
    set match=true
    set TOPIC_IN=%2
    shift
)
if [%key%]==[-o] (
    set match=true
    set TOPIC_OUT=%2
    shift
)
if [%key%]==[-d] (
    set match=true
    set PUB_DATA=%2
    shift
)
if [%key%]==[-r] (
    set match=true
    set PUB_REPLY=true
)
if [%key%]==[-e] (
    set match=true
    set PUB_EVAL=true
)
if [%key%]==[-u] (
    set match=true
    set USERACCT=%2
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
if ["%FNPATH%"]==[""] (
  set FNPATH=.
) else (
  if "%FNPATH:~-1%"=="\" (
    set FNPATH=%FNPATH:~0,-1%
  )
)
if [%FUNCTION%]==[] (
  if ["%FNPATH%"]==["."] (
    for %%a in (.) do set FUNCTION=%%~na
  ) else (
    for %%a in ("%FNPATH%") do set FUNCTION=%%~na
  )
)
if [%VERSION%]==[] set VERSION=0.0.1
if [%CONTAINER%]==[] set CONTAINER=sidecar
if [%USERACCT%]==[] set USERACCT=%username%

:: execute the commands
if [%_COMMAND%]==[init] (
    call :init
    exit /B %ERRORLEVEL%
)

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

if [%_COMMAND%]==[update] (
    call :update
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

if [%_COMMAND%]==[publish] (
    call :publish
    exit /B %ERRORLEVEL%
)

if [%_COMMAND%]==[version] (
    call :version
    exit /B %ERRORLEVEL%
)

exit /B 1

:: commands
:init
  :: assume the file extension based on language - might be overridden as necessary later
  if [%FNLANG%]==[] set FNEXT=.sh
  if [%FNLANG%]==[shell] set FNEXT=.sh
  if [%FNLANG%]==[java] set FNEXT=.jar
  if [%FNLANG%]==[node] set FNEXT=.js
  if [%FNLANG%]==[python] set FNEXT=.py
  :: check --file argument
  if exist "%FNPATH%" (
    set _exists=true
  ) else (
    echo "%FNPATH%": No such file or directory
    exit /B 1
  )
  if [%_exists%]==[true] (
    for %%p IN ("%FNPATH%") do if exist %%~sp\nul set FNDIR=%%~sp
  )
  if [%_exists%]==[true] (
    if [%FNDIR%]==[] (
      :: FNPATH is a file
      for /F "delims=" %%i in ("%FNPATH%") do set FNFILE=%%~ni%%~xi
      for /F "delims=" %%i in ("%FNPATH%") do set FNEXT=%%~xi
      for /F "delims=" %%i in ("%FNPATH%") do set FNDIR=%%~dspi
    ) else (
      :: FNPATH is a directory
      if exist %FNDIR%\%FUNCTION%.sh set FNEXT=.sh& set FNFILE=%FUNCTION%.sh
      if exist %FNDIR%\%FUNCTION%.jar set FNEXT=.jar& set FNFILE=%FUNCTION%.jar
      if exist %FNDIR%\%FUNCTION%.js set FNEXT=.js& set FNFILE=%FUNCTION%.js
      if exist %FNDIR%\%FUNCTION%.py set FNEXT=.py& set FNFILE=%FUNCTION%.py
    )
  )
  :: override file path and extension if --artifact provided
  if [%FNARTIFACT%]==[] (
    set ARTRELPATH=%FNFILE%
    set ARTRELDIR=.
    set _artifact=not-set
  ) else (
    set _artifact=false
    if exist %FNARTIFACT% set _artifact=true
  )
  if [%_artifact%]==[false] (
      echo Artifact %FNARTIFACT%: No such file
      exit /B 1
  ) 
  if [%_artifact%]==[true] (
    set ARTRELPATH=%FNARTIFACT%
    set ARTRELDIR="."
    for /F %%i in ("%FNARTIFACT%") do set FNFILE=%%~ni%%~xi
    for /F %%i in ("%FNARTIFACT%") do set FNEXT=%%~xi
  )
  :: figure out language if not provided in --language arg
  if [%FNLANG%]==[] (
    if "%FNEXT%"==".sh" set FNLANG=shell
    if "%FNEXT%"==".jar" set FNLANG=java
    if "%FNEXT%"==".js" set FNLANG=node
    if "%FNEXT%"==".py" set FNLANG=python
  )
  :: figure out protocol if not provided in --protocol arg
  if [%FNPROT%]==[] (
    if "%FNEXT%"==".sh" set FNPROT=stdio
    if "%FNEXT%"==".jar" set FNPROT=http
    if "%FNEXT%"==".js" set FNPROT=http
    if "%FNEXT%"==".py" set FNPROT=stdio      
  )
  :: default the topic to function name
  if [%TOPIC_IN%]==[] set TOPIC_IN=%FUNCTION%
  :: ready to initialize
  :: but first look for function source
  if exist "%ARTRELPATH%" (
    echo Initializing %FNLANG% function %FUNCTION%
  ) else (
    echo Source file %FNFILE% not found, not able to initialize %FNLANG% function %FUNCTION%
    exit /B 1
  )
  set FNDOCKER=%FNDIR%\Dockerfile
  if exist %FNDOCKER% (
    echo Docker file already exists in %FNDIR%
  ) else (
    call :write_dockerfile
  )
  set FNDEF=%FNDIR%\%FUNCTION%-function.yaml
  if exist %FNDEF% (
    echo Function definition file already exists in %FNDIR%
  ) else (
    call :write_function_yaml
  )
  set FNTOPICS=%FNDIR%\%FUNCTION%-topics.yaml
  if exist %FNTOPICS% (
    echo Topics definition file already exists in %FNDIR%
  ) else (
    call :write_topics_yaml
  )
  exit /B 1

:build
  echo Building %FUNCTION% %FNPATH% %VERSION%
  docker build -t %USERACCT%/%FUNCTION%:%VERSION% %FNPATH%
  if "$DOCKERPUSH"=="true" docker push %USERACCT%/%FUNCTION%:%VERSION%
  exit /B %ERRORLEVEL%

:create
  echo Creating function %FUNCTION% version %VERSION%
  call :init
  set _dir=false
  for %%p IN ("%FNPATH%") do if exist %%~sp\nul set _dir=true
  if "%_dir%"=="false" (
    for /F "delims=" %%i in ("%FNPATH%") do set FNPATH=%%~dspi
  )
  call :build
  call :apply
  exit /B %ERRORLEVEL%

:apply
  echo Applying %FNPATH% resource[s]
  kubectl apply -f %FNPATH%
  exit /B %ERRORLEVEL%

:update
  set _dir=false
  for %%p IN ("%FNPATH%") do if exist %%~sp\nul set _dir=true
  if "%_dir%"=="false" (
    for /F "delims=" %%i in ("%FNPATH%") do set FNPATH=%%~dspi
  )
  set FNDOCKER=%FNPATH%\Dockerfile
  if exist %FNDOCKER% (
    echo Updating %FNPATH% resource[s]
    kubectl delete -f %FNPATH%
    call :build
    call :apply
  ) else (
    echo Resource files not found in %FNPATH% directory, not able to update function %FUNCTION%
    exit /B 1
  )
  exit /B %ERRORLEVEL%

:delete
  echo Deleting %FUNCTION% resource[s]
  if "%DELETEALL%"=="true" (
    if Not exist %FNPATH%\*.yaml (
      echo No resource definitions found in the %FNPATH% directory"
    ) else (
      kubectl delete -f %FNPATH%
    )
  ) else (
    kubectl delete function %FUNCTION%
  )
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

:publish
  call :capture_instances svc "http-gateway"
  if "%svc%"=="[]" (
    echo Unable to locate the http-gateway
    exit /B 1
  )

  call :capture_jsonpath svc_type "http-gateway" "{.items[0].spec.type}"
  if [%svc_type%]==[NodePort] (
    call :capture_cmd address "minikube ip"
    call :capture_jsonpath port "http-gateway" "{.items[0].spec.ports[?(@.name == 'http')].nodePort}"
    goto do_curl
  )

  call :capture_jsonpath address "http-gateway" "{.items[0].status.loadBalancer.ingress[0].ip}"
  if [%address%]==[] (
    echo External IP is not yet available, try in a few ...
    exit /B 1
  )
  call :capture_jsonpath port "http-gateway" "{.items[0].spec.ports[?(@.name == 'http')].port}"

  :do_curl
  if [%TOPIC_IN%]==[] set TOPIC_IN=%FUNCTION%
  if [%PUB_COUNT%]==[] (
      set count=1
  ) else (
      set count=%PUB_COUNT%
  )
  if [%PUB_PAUSE%]==[] (
      set pause=0
  ) else (
      set pause=%PUB_PAUSE%
  )
  for /L %%i in (1,1,%count%) do call :do_post_content %%i
  echo.

  exit /B %ERRORLEVEL%

:version
  echo riff version %RIFF_VERSION%
  exit /B 0

:write_dockerfile
  if "%FNLANG%"=="shell" (
    echo FROM projectriff/%FNLANG%-function-invoker:%RIFF_VERSION% >> %FNDOCKER%
    echo ARG FUNCTION_URI="${FNFILE}" >> %FNDOCKER%
    echo ENV FUNCTION_URI ${FUNCTION_URI} >> %FNDOCKER%
    echo ADD %ARTRELPATH% / >> %FNDOCKER%
  )
  if "%FNLANG%"=="java" (
    echo FROM projectriff/%FNLANG%-function-invoker:%RIFF_VERSION% >> %FNDOCKER%
    echo ARG FUNCTION_JAR=/functions/%FNFILE% >> %FNDOCKER%
    echo ARG FUNCTION_CLASS=%FNHANDLER% >> %FNDOCKER%
    echo ENV FUNCTION_URI file://${FUNCTION_JAR}?handler=${FUNCTION_CLASS} >> %FNDOCKER%
    echo ADD %ARTRELPATH% ${FUNCTION_JAR} >> %FNDOCKER%
  )
  if "%FNLANG%"=="node" (
    echo FROM projectriff/%FNLANG%-function-invoker:%RIFF_VERSION% >> %FNDOCKER%
    echo ENV FUNCTION_URI /functions/%FNFILE% >> %FNDOCKER%
    echo ADD %ARTRELPATH% ${FUNCTION_URI} >> %FNDOCKER%
  )
  if "%FNLANG%"=="python" (
    echo FROM projectriff/python2-function-invoker:%RIFF_VERSION% >> %FNDOCKER%
    echo ARG FUNCTION_MODULE=%FNFILE% >> %FNDOCKER%
    echo ARG FUNCTION_HANDLER=process >> %FNDOCKER%
    echo ENV FUNCTION_URI file:///${FUNCTION_MODULE}?handler=${FUNCTION_HANDLER} >> %FNDOCKER%
    echo ADD %ARTRELPATH% / >> %FNDOCKER%
    echo ADD %ARTRELDIR%/requirements.txt / >> %FNDOCKER%
    echo RUN pip install --upgrade pip ^&^& pip install -r /requirements.txt >> %FNDOCKER%
  )
  exit /B %ERRORLEVEL%

:write_function_yaml
  echo apiVersion: projectriff.io/v1 >> %FNDEF%
  echo kind: Function >> %FNDEF%
  echo metadata: >> %FNDEF%
  echo   name: %FUNCTION% >> %FNDEF%
  echo spec: >> %FNDEF%
  echo   protocol: %FNPROT% >> %FNDEF%
  echo   input: %TOPIC_IN% >> %FNDEF%
  if Not "%TOPIC_OUT%"=="" (
    echo   output: %TOPIC_OUT% >> %FNDEF%
  )
  echo   container: >> %FNDEF%
  echo     image: %USERACCT%/%FUNCTION%:%VERSION% >> %FNDEF%
  exit /B %ERRORLEVEL%

:write_topics_yaml
  echo apiVersion: projectriff.io/v1 >> %FNTOPICS%
  echo kind: Topic >> %FNTOPICS%
  echo metadata: >> %FNTOPICS%
  echo   name: %TOPIC_IN% >> %FNTOPICS%
  echo spec: >> %FNTOPICS%
  echo   partitions: 1 >> %FNTOPICS%
  if Not "%TOPIC_OUT%"=="" (
    echo --- >> %FNTOPICS%
    echo apiVersion: projectriff.io/v1 >> %FNTOPICS%
    echo kind: Topic >> %FNTOPICS%
    echo metadata: >> %FNTOPICS%
    echo   name: %TOPIC_OUT% >> %FNTOPICS%
    echo spec: >> %FNTOPICS%
    echo   partitions: 1 >> %FNTOPICS%
  )
  exit /B %ERRORLEVEL%

:do_post_content
if "%PUB_EVAL%"=="true" (
  call :capture_echo _message %PUB_DATA% %1
) else (
  set _message=%PUB_DATA%
)
if "%PUB_REPLY%"=="true" (
  curl -H "Content-Type: text/plain" -X POST http://%address%:%port%/requests/%TOPIC_IN% -d %_message%
) else (
  curl -H "Content-Type: text/plain" -X POST http://%address%:%port%/messages/%TOPIC_IN% -d %_message%
)
timeout /t %pause% /nobreak > NUL
exit /B %ERRORLEVEL%

:capture_echo
set i=%3
for /f "tokens=* usebackq" %%f in (`echo %2`) do (
  set %1=%%f
)
exit /B %ERRORLEVEL%

:capture_cmd
for /f "tokens=* usebackq" %%f in (`%~2`) do (
  set %1=%%f
)
exit /B %ERRORLEVEL%

:capture_instances
set _tmp_file=%tmp%riff-%random%.txt
kubectl get svc -l component=%~2 -o jsonpath="{.items}" >  %_tmp_file%
for /f "delims= tokens=1" %%x in (%_tmp_file%) do set %1=%%x
del %_tmp_file%
exit /B %ERRORLEVEL%

:capture_jsonpath
set _tmp_file=%tmp%\riff-%random%.txt
set _kcmd=kubectl get svc -l component=%~2 -o jsonpath=%3
%_kcmd% > %_tmp_file%
for /f "delims= tokens=1" %%x in (%_tmp_file%) do set %1=%%x
del %_tmp_file%
exit /B %ERRORLEVEL%

:: help texts
:print_usage
echo.
echo riff is for functions
echo.
echo version %RIFF_VERSION%
echo.
echo Commands:
echo   init         Initialize a function
echo   build        Build a function container
echo   apply        Apply function resource definitions
echo   create       Create function resources
echo   update       Upfate function resources
echo   delete       Delete function resources
echo   list         List current function resources
echo   logs         Show logs for a function resource
echo   publish      Publish data to a topic using the http-gateway
echo   version      Display the riff version
echo.
echo   Use "riff <command> --help" for more information about a given command.
echo.
exit /B 0

:print_init_usage
echo.
echo Initialize the function based on the function source code specified as the filename, using the name
echo   and version specified for the function image repository and tag.
echo.
echo Usage:
echo.
echo   riff init -u ^<useraccount^> -n ^<name^> -v ^<version^> -f ^<source^> -l ^<language^> -p ^<protocol^> -i ^<input-topic^> -o ^<output-topic^> [-a ^<artifact^>] [--handler ^<handler-name^>] [--push]
echo.
echo Options:
echo.
echo   -u, --useraccount: the Docker user account to be used for the image repository (defaults to current OS username)
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -v, --version: the version of the function (defaults to 0.0.1)
echo   -f, --filename: filename or directory to be used for the function resources, 
echo                   if a file is specified then the file's directory will be used 
echo                   (defaults to the current directory)
echo   -l, --language: the language used for the function source
echo                   (defaults to filename extension language type or 'shell' if directory specified)
echo   -p, --protocol: the protocol to use for function invocations
echo                   (defaults to 'stdio' for shell and python, to 'http' for java and node)
echo   -i, --input: the name of the input topic (defaults to function name)
echo   -o, --output: the name of the output topic (no default, only created if specified)
echo   -a, --artifact: path to the function artifact, source code or jar file
echo                   (defaults to function name with extension appended based on language: 
echo                       '.sh' for shell, '.jar' for java, '.js' for node and '.py' for python)
echo   --handler: the name of the handler, for Java it is the fully qualified class name of the Java class where the function is defined
echo   --riff-version: the version of riff to use when building containers
echo   --push: push the image to Docker registry
echo.
exit /B 0

:print_build_usage
echo.
echo Build the function based on the code available in the path directory, using the name
echo   and version specified for the image that is built.
echo.
echo Usage:
echo.
echo   riff build -n ^<name^> -v ^<version^> -f ^<path^> [--push]
echo.
echo Options:
echo.
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -v, --version: the version of the function (defaults to 0.0.1)
echo   -f, --filename: filename, directory, or URL for the code or resource (defaults to the current directory)
echo   --push: push the image to Docker registry
echo.
exit /B 0

:print_create_usage
echo.
echo Create the resource[s] for the function based on the function source code specified as the filename, using the name
echo   and version specified for the function image repository and tag. Build the function and deploy it.
echo.
echo Usage:
echo.
echo   riff create -u ^<useraccount^> -n ^<name^> -v ^<version^> -f ^<source^> -l ^<language^> -p ^<protocol^> -i ^<input-topic^> -o ^<output-topic^> [-a ^<artifact^>] [--handler ^<handler-name^>] [--push]
echo.
echo Options:
echo.
echo   -u, --useraccount: the Docker user account to be used for the image repository (defaults to current OS username)
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -v, --version: the version of the function (defaults to 0.0.1)
echo   -f, --filename: filename or directory to be used for the function resources, 
echo                   if a file is specified then the file's directory will be used 
echo                   (defaults to the current directory)
echo   -l, --language: the language used for the function source
echo                   (defaults to filename extension language type or 'shell' if directory specified)
echo   -p, --protocol: the protocol to use for function invocations
echo                   (defaults to 'stdio' for shell and python, to 'http' for java and node)
echo   -i, --input: the name of the input topic (defaults to function name)
echo   -o, --output: the name of the output topic (no default, only created if specified)
echo   -a, --artifact: path to the function artifact, source code or jar file
echo                   (defaults to function name with extension appended based on language: 
echo                       '.sh' for shell, '.jar' for java, '.js' for node and '.py' for python)
echo   --handler: the name of the handler, for Java it is the fully qualified class name of the Java class where the function is defined
echo   --riff-version: the version of riff to use when building containers
echo   --push: push the image to Docker registry
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

:print_update_usage
echo.
echo This feature is not yet available for Windows. 
echo.
echo Contributions are welcome, refer to https://github.com/projectriff/riff/issues/235
echo.
exit /B 0

:print_delete_usage
echo.
echo Delete the resource[s] for the function specified.
echo.
echo Usage:
echo.
echo   riff delete -n ^<name^>
echo     or
echo   riff delete -f ^<path^> --all
echo.
echo Options:
echo.
echo   -n, --name: the name of the function (defaults to the name of the current directory)
echo   -f, --filename: filename, directory, or URL for the resource definition[s] (defaults to the current directory)
echo   --all: delete all resources including topics, not just the function resource
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

:print_publish_usage
echo.
echo Publish data to a topic using the http-gateway.
echo.
echo Usage:
echo.
echo   riff publish -i <input-topic> -d <data> [-r] [--count <count>] [--pause <pause>]
echo.
echo Options:
echo.
echo   -i, --input: the name of the input topic (defaults to the name of the current directory)
echo   -d, --data: the data to post to the http-gateway using the input topic
echo   -r, --reply: wait for a reply containing the results of the function execution
echo   -e, --eval: evaluate the data and substitute variables
echo               (e.g. you can use %%%%i%%%% to capture the iteration when using --count)
echo   --count: the number of times to post the data (defaults to 1)
echo   --pause: the number of seconds to wait between postings (defaults to 0)
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
