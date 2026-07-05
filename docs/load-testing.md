Stress Test for User Apps ----->>>>>>||>>>>

StressFree Test 01:
1000 concurrent requests -|>
 HTTP
    http_req_duration..............: avg=238.37ms min=149.27ms med=220.51ms max=3.3s  p(90)=238.64ms p(95)=295.81ms
      { expected_response:true }...: avg=238.37ms min=149.27ms med=220.51ms max=3.3s  p(90)=238.64ms p(95)=295.81ms
    http_req_failed................: 0.00% 0 out of 96694
    http_reqs......................: 96694 797.607429/s

    EXECUTION
    iteration_duration.............: avg=1.24s    min=1.15s    med=1.22s    max=5.78s p(90)=1.23s    p(95)=1.32s   
    iterations.....................: 96694 797.607429/s
    vus............................: 350   min=350        max=1000
    vus_max........................: 1000  min=1000       max=1000

    NETWORK
    data_received..................: 66 MB 545 kB/s
    data_sent......................: 14 MB 114 kB/s


stressFree-test 02:
1500 concurrent request -|>


  █ TOTAL RESULTS 

    checks_total.......: 142014  1171.472353/s
    checks_succeeded...: 100.00% 142014 out of 142014
    checks_failed......: 0.00%   0 out of 142014

    ✓ status is 200

    HTTP
    http_req_duration..............: avg=250.71ms min=188.39ms med=221.49ms max=3.36s  p(90)=243.12ms p(95)=332.39ms
      { expected_response:true }...: avg=250.71ms min=188.39ms med=221.49ms max=3.36s  p(90)=243.12ms p(95)=332.39ms
    http_req_failed................: 0.00%  0 out of 142014
    http_reqs......................: 142014 1171.472353/s

    EXECUTION
    iteration_duration.............: avg=1.27s    min=1.18s    med=1.22s    max=11.72s p(90)=1.24s    p(95)=1.41s   
    iterations.....................: 142014 1171.472353/s
    vus............................: 468    min=468         max=1500
    vus_max........................: 1500   min=1500        max=1500

    NETWORK
    data_received..................: 97 MB  800 kB/s
    data_sent......................: 20 MB  168 kB/s


stressFree-test 03:
3000 concurrent request -|>
Uncontrolled Stree Test 
Finding: system saturated under extreme load; controlled baseline testing required
Not good


Stress Test for forge-itself ----->>>>>>||>>>>
# API Load Test 001

## Goal
Evaluate Forge API performance under concurrent read traffic.

## Configuration

- Tool: k6
- Duration: 2m 30s
- Target: 900 RPS

| Endpoint | RPS |
|----------|----:|
| /api/health | 500 |
| /api/projects | 300 |
| /api/projects/:id/url | 100 |

## Results

| Metric | Value |
|---------|------:|
| Achieved RPS | 676 |
| Avg Latency | 43s |
| P99 | 33s |
| Failed Requests | 6.2% |
| Dropped Iterations | 6900 |

## Findings

- Target throughput not achieved.
- Backend saturated under concurrent read traffic.
- Database-backed endpoints showed significant latency.
- Further profiling required to isolate the primary bottleneck.

## Next Experiment

- Backend Replicas: **1 → 3**
- Re-run identical workload.
- Compare latency, throughput and failure rate.


