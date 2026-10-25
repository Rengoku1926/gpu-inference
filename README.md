# gpu-inference

A small serverless GPU inference platform: register a model, call an endpoint, get streamed tokens, with nothing running when nobody calls.

**Commitment (through April 2027):** serve Qwen2.5-7B on a single 24 GB GPU with p50 TTFT under 200 ms, p99 TTFT under 1 s, and p99 end-to-end latency under 2 s for 256 output tokens at 32 concurrent streams, with cold start under 30 s.
