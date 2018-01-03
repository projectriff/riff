@echo off
:: ##########################################################################
::  riff demo command for Windows to send request messages
:: ##########################################################################

:: set local scope for the variables with windows NT shell
if [%OS%]==[Windows_NT] setlocal

call :capture_instances svc "http-gateway"
if "%svc%"=="[]" (
  echo Unable to locate the http-gateway
  exit /B 1
)

call :capture_jsonpath svc_type "http-gateway" "{.items[0].spec.type}"
if [%svc_type%]==[NodePort] (
  call :capture_cmd address "minikube ip"
  call :capture_jsonpath port "http-gateway" "{.items[0].spec.ports[?(@.name == 'http')].nodePort}"
) else (
  call :capture_jsonpath address "http-gateway" "{.items[0].status.loadBalancer.ingress[0].ip}"
  if [%address%]==[] (
    echo External IP is not yet available, try in a few ...
    exit /B 1
  )
  call :capture_kubectl_jsonpath_cmd port "http-gateway" "{.items[0].spec.ports[?(@.name == 'http')].port}"
)

curl -H "Content-Type: text/plain" -X POST http://%address%:%port%/requests/%~1 -d "%~2"
echo.

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
