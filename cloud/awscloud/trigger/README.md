trigger 
========

```trigger``` provides a helper to write lambda based workers.

### Overview

To use this package, you'll need to construct your event stream as follows: 

```text
App ---> SNS ---> SQS
          |        ^
          |        |
          v        |
        Lambda ----
```

1. The app fires an event into SNS
2. An SQS subscriber receives the event (ridiculously reliable operation)
3. A lambda subscriber also receives the event (not nearly so reliable)
4. The lambda function queries SQS for an outstanding events and processes
5. In the case the event failed, the even will remain in SQS and be picked up by the CloudWatch handler (shown below)

### Handling Errors

```text
CloudWatch ---> SNS ---> Lambda ---> SQS
```

In case, the initial lambda call fails, we need a way to reprocess the records.  Here we'll use scheduled CloudWatch 
events to trigger another firing of our lambda function.

1. CloudWatch fires a scheduled event into SNS
    * To minimize the number of CloudWatch scheduled events we need, we'll use a generic topic, every-1min, for example
2. A lambda subscriber receives the event
    * This will be the same instance as used by the initial processing
3. The lambda function queries SQS for an outstanding events and processes
4. In the case of error, the event remains on the queue and will be processed again
