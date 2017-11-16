FROM sk8s/python2-function-invoker:0.0.1-SNAPSHOT
ARG FUNCTION_MODULE=sentiment_service.py
ARG FUNCTION_HANDLER=process
ADD ./sentiment_service.py /
ADD ./requirements.txt /
RUN  pip install --upgrade pip && pip install -r /requirements.txt
ENV FUNCTION_URI file:///${FUNCTION_MODULE}?${FUNCTION_HANDLER}
