"""The terminal event of a failed run on the FastAPI endpoint.

A failure inside the agent still exits the pyagentspec ``Trace``, whose
shutdown makes the span processor emit RUN_FINISHED; the endpoint then
reports the exception as RUN_ERROR. The client therefore saw the run finish
successfully and then fail.
"""

import json
import unittest

from fastapi import FastAPI
from fastapi.testclient import TestClient
from pyagentspec.tracing.spans.span import Span
from pyagentspec.tracing.trace import Trace

from ag_ui_agentspec.agentspec_tracing_exporter import AgUiSpanProcessor
from ag_ui_agentspec.endpoint import add_agentspec_fastapi_endpoint

_INPUT = {
    "threadId": "thread-1",
    "runId": "run-1",
    "messages": [],
    "state": {},
    "tools": [],
    "context": [],
    "forwardedProps": {},
}


class _Agent:
    """Runs the way AgentSpecAgent.run does: a Trace around an invoke span."""

    def __init__(self, fail: bool):
        self.fail = fail
        self.processors = [AgUiSpanProcessor(runtime="langgraph")]

    async def run(self, input_data):
        async with Trace(name="ag-ui run wrapper", span_processors=self.processors):
            async with Span(name="invoke_graph"):
                if self.fail:
                    raise RuntimeError("model provider returned 500")


def _event_types(agent):
    app = FastAPI()
    add_agentspec_fastapi_endpoint(app, agent, "/")
    body = TestClient(app).post("/", json=_INPUT).text
    return [
        json.loads(line[len("data:"):])["type"]
        for line in body.splitlines()
        if line.startswith("data:")
    ]


class TestEndpointRunError(unittest.TestCase):
    def test_failed_run_ends_with_run_error_only(self):
        # Guards against RUN_FINISHED (successful completion) preceding the
        # RUN_ERROR of the same run.
        types = _event_types(_Agent(fail=True))
        self.assertEqual(types, ["RUN_STARTED", "RUN_ERROR"])

    def test_successful_run_still_finishes(self):
        types = _event_types(_Agent(fail=False))
        self.assertEqual(types, ["RUN_STARTED", "RUN_FINISHED"])


if __name__ == "__main__":
    unittest.main()
