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

    
Uncontrolled Stress Test 02:
57.1k requests, 42.3% failures, p95 4s, p99 6s.
Finding: system saturated under extreme load; controlled baseline testing required
