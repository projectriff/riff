FROM projectriff/python3-function-invoker:0.0.5-snapshot
ARG FUNCTION_MODULE=echo.py
ARG FUNCTION_HANDLER=matching-invoker-with-output-topic
ADD ./echo.py /
ENV FUNCTION_URI file:///${FUNCTION_MODULE}?handler=${FUNCTION_HANDLER}
