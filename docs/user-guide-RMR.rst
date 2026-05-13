..
..  Copyright (c) 2019 AT&T Intellectual Property.
..  Copyright (c) 2019 Nokia.
..
..  Licensed under the Creative Commons Attribution 4.0 International
..  Public License (the "License"); you may not use this file except
..  in compliance with the License. You may obtain a copy of the License at
..
..    https://creativecommons.org/licenses/by/4.0/
..
..  Unless required by applicable law or agreed to in writing, documentation
..  distributed under the License is distributed on an "AS IS" BASIS,
..  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
..
..  See the License for the specific language governing permissions and
..  limitations under the License.
..

User-Guide (old)
================

.. contents::
   :depth: 3
   :local:

``RMR based interface between xApp and Subscription Manager which is explained in this User-Guide``
``is deprecated and will be removed in the future. See second User-Guide which explains new REST``
``based interface.``

Overview
--------
Subscription Manager is a basic platform service in RIC. It is responsible for managing E2 subscriptions from xApps to the
Radio Access Network (RAN).

xApp can subscribe and unsubscribe messages from gNodeB through Subscription Manager. Subscription Manager manages the subscriptions
and message routing of the subscribed messages between E2 Termination and xApps. If one xApp has already made a subscription and then
another xApp initiates identical subscription, Subscription Manager does not forward the subscription to gNodeB but merges the
subscriptions internally. In merge case Subscription Manager just updates the message routing information to Routing Manager and
sends response to xApp.

There can be only one ongoing RIC Subscription or RIC Subscription Delete procedure towards RAN at any time. That is because Subscription
Manager is able to merge new subscriptions only which those it has already received successful response from RAN. Subscriptions
and delete subscriptions are therefore queued in Subscription Manager. Subscription Manager may need to do reties during subscribe or
unsubscribe procedure. This needs to be considered when retries are implemented in xApp side, as it can increase completion time of the
procedure. xApp's retry delay should not be too short.

    .. image:: images/PlaceInRICSoftwareArchitecture_RMR.png
      :width: 600
      :alt: Place in RIC's software architecture picture

Architecture
------------

  * Message routing

      Subscribed messages from RAN are transported to RIC inside RIC Indication message. RIC Indication message is transported to xApp
      inside RMR message, in Payload field of the RMR message. RMR message is routed to xApp based on SubId field (subscription id) in
      the RMR header. Same routing mechanism is used also for response messages from Subscription Manager to xApp. Subscription Manager is
      not able to respond to xApp if route for subscription has not been created.

      When xApp sends message to Subscription Manager it sets -1 in SubId field in the RMR header. It means that messages are routed based
      on message type (Mtype filed in RMR header). RIC Subscription Request and RIC Subscription delete Request messages are pre configured
      to be routed to Subscription Manager.

      Subscription Manager allocates unique RIC Request Sequence Number for every subscription during Subscription procedure. Subscription
      Manager replaces existing ASN.1 decoded RIC Request Sequence Number in the RIC Subscription Request message allocated by the xApp.
      That sequence number (subscription id) is then used for the subscription in RIC and RAN as long the subscription lives. xApp gets
      the sequence number in RIC Subscription Response message from Subscription manager based on the message types.
      
      Subscribed messages are routed to xApps based on sequence number. Sequence number is placed in the SubId field of the RMR message
      header when E2 Termination sends the subscribed message to xApp. When xApp wants to delete the subscription, the same sequence number
      must be included in the ASN.1 encoded RIC Subscription Delete Request message sent to Subscription Manager.

      Subscription Manager allocates and substitutes RIC Requestor ID for E2 interface communication. Currently the ID value is always 123.
      RAN gets the Request of the xApp who makes the first subscription. RAN uses Subscription Manager allocated Requestor ID in all RIC
      Indication messages it sends to RIC for the subscription.  Subscription Manager returns the same RIC Requestor ID in response message
      to xApp as was in the request. In merge case subscription is created only for the first requestor.

      TransactionId (Xid) in RMR message header is used to track messages between xApp and Subscription Manager. xApp allocates it. Subscription
      Manager returns TransactionId received from xApp in response message to xApp. xApp uses it to map response messages to request messages
      it sends.

  * Subscription procedure
      
    * Successful case

      xApp sends RIC Subscription Request message to Subscription Manager. Subscription Manager validates request types in the message and sends
      route create to Routing Manager over REST interface. When route is created successfully Subscription Manager forwards request to E2
      Termination. When RIC Subscription Response arrives from E2 Termination Subscription Manager forwards it to xApp.
      
      Subscription Manager supervises route create and RIC Subscription Request with a timer.

      RIC Indication messages which are used to transport subscribed messages from RAN are routed from E2 Termination to xApps
      directly using the routes created during Subscription procedure.

      Subscription Manager supports REPORT, POLICY and INSERT type subscriptions (RICActionTypes). CONTROL is not supported. POLICY type
      subscription can be updated. In update case signaling sequence is the same as above, except route is not created to Routing manager.
      xApp uses initially allocated TransactionId and RIC Request Sequence Number in update case. Route in POLICY type subscription case is needed
      only that Subscription Manager can send response messages to xApp. RIC Subscription Request message contains list of RICaction-ToBeSetup-ItemIEs.
      The list cannot have REPORT, POLICY, INSERT action types at the same time. Subscription Manager checks actions types in the message.
      If both action types is found the message is dropped.


    .. image:: images/Successful_Subscription_ASN.1.png
      :width: 600
      :alt: Successful subscription picture


    * Failure case

      In failure case Subscription Manager checks the failure cause and acts based on that. If failure cause is "duplicate" Subscription
      Manager sends delete to RAN and then resends the same subscription. If failure cause is such that Subscription manager cannot do
      anything to fix the problem, it sends delete to RAN and sends RIC Subscription Failure to xApp. Subscription Manager may retry RIC
      Subscription Request and RIC Subscription Delete messages also in this case before it responds to xApp.

    .. image:: images/Subscription_Failure_ASN.1.png
      :width: 600
      :alt: Subscription failure picture

    * Timeout in Subscription Manager

     In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Request to RAN. If there is no response
      after retry, Subscription Manager shall NOT send any response to xApp. xApp may retry RIC Subscription Request, if it wishes to do so.
      Subscription Manager does no handle the retry if Subscription Manager has ongoing subscription procedure for the same subscription.
      Subscription just drops the request.

    * Timeout in xApp

      xApp may resend the same request if there is no response in expected time. If xApp resends the same request while processing of previous
      request has not been completed in Subscription Manager then Subscription Manager drops the new request, makes a log writing and continues
      processing previous request.

    .. image:: images/Subscription_Timeout_ASN.1.png
      :width: 600
      :alt: Subscription timeout picture

  * Subscription delete procedure

    * Successful case

      xApp sends RIC Subscription Delete Request message to Subscription Manager. xApp must use the same RIC Request Sequence Number which
      it received in RIC Subscription Response message when subscription is deleted. When Subscription Manager receives RIC Subscription
      Delete Request message, Subscription Manager first forwards the request to E2 Termination. When RIC Subscription Delete Response arrives
      from E2 Termination to Subscription Manager, Subscription Manager forwards it to xApp and then request route deletion from Routing Manager.
      
      Subscription Manager supervises RIC Subscription Deletion Request and route delete with a timer.

    .. image:: images/Successful_Subscription_Delete_ASN.1.png
      :width: 600
      :alt: Successful subscription delete picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager always responds with RIC Subscription Delete Response to xApp.

    .. image:: images/Subscription_Delete_Failure_ASN.1.png
      :width: 600
      :alt: Subscription delete failure picture

    * Timeout in Subscription Manager

      In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Delete Request to RAN. If there is no
      response after retry, Subscription Manager responds to xApp with RIC Subscription Delete Response.

    * Timeout in xApp

      xApp may resend the same request if there is no response in expected time. If xApp resends the same request while processing of previous
      request has not been completed in Subscription Manager then Subscription Manager drops the new request, makes a log writing and continues
      processing previous request.

    .. image:: images/Subscription_Delete_Timeout_ASN.1.png
      :width: 600
      :alt: Subscription delete timeout picture

    * Unknown subscription

      If Subscription Manager receives RIC Subscription Delete Request for a subscription which does not exist, Subscription Manager cannot respond
      to xApp as there is no route for the subscription.

  * Subscription merge procedure

    * Successful case

      xApp sends RIC Subscription Request message to Subscription Manager. Subscription Manager checks is the Subscription mergeable. If not,
      Subscription Manager continues with normal Subscription procedure. If Subscription is mergeable, Subscription Manager sends route create
      to Routing Manager and then responds with RIC Subscription Response to xApp.
      
      Route create is supervised with a timer.

      Merge for REPORT type subscription is possible if Action Type and Event Trigger Definition of subscriptions are equal.

      ``Only REPORT type subscriptions can be be merged.``

    .. image:: images/Successful_Subscription_Merge_ASN.1.png
      :width: 600
      :alt: Successful subscription merge picture

    * Failure case

      Failure case is basically the same as in normal subscription procedure. Failure can come only from RAN when merge is not yet done.
      If error happens during route create Subscription Manager drops the RIC Subscription Request message and xApp does not get any response.

    * Timeout in Subscription Manager

      Timeout case is basically the same as in normal subscription procedure but timeout can come only in route create during merge operation.
      If error happens during route create, Subscription Manager drops the RIC Subscription Request message and xApp does not get any response.

    * Timeout in xApp

      xApp may resend the same request if there is no response in expected time. If xApp resends the same request while processing of previous
      request has not been completed in Subscription Manager then Subscription Manager drops the new request, makes a log writing and continues
      processing previous request.

  * Subscription delete merge procedure

    * Successful case

      xApp sends RIC Subscription Delete Request message to Subscription Manager. If delete concerns merged subscription, Subscription Manager
      responds with RIC Subscription Delete Response to xApp and then sends route delete request to Routing manager.
      
      Subscription Manager supervises route delete with a timer.

    .. image:: images/Successful_Subscription_Delete_Merge_ASN.1.png
      :width: 600
      :alt: Successful subscription delete merge picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager responds with RIC Subscription Delete Response message to xApp.

    * Timeout in Subscription Manager

      Timeout can only happen in route delete to Routing manager. Subscription Manager responds with RIC Subscription Delete Response message to xApp.

    * Timeout in xApp

      xApp may resend the same request if there is no response in expected time. If xApp resends the same request while processing of previous
      request has not been completed in Subscription Manager then Subscription Manager drops the new request, makes a log writing and continues
      processing previous request.

  * Unknown message

     If Subscription Manager receives unknown message, Subscription Manager drops the message.

  * xApp restart

    When xApp is restarted for any reason it may resend subscription requests for subscriptions which have already been subscribed. If REPORT or INSERT type
    subscription already exists and RMR endpoint of requesting xApp is attached to subscription then successful response is sent to xApp directly without
    updating Routing Manager and BTS. If POLICY type subscription already exists, request is forwarded to BTS and successful response is sent to xApp.
    BTS is expected to accept duplicate POLICY type requests. In restart IP address of the xApp may change but domain service address name does not.
    RMR message routing uses domain service address name.

  * Subscription Manager restart

    Subscription Manager stores successfully described subscriptions from db (SDL). Subscriptions are restored from db in restart. For subscriptions which
    were not successfully completed, Subscription Manager sends delete request to BTS and removes routes created for those. Restoring subscriptions from
    db can be disable via submgr-config.yaml file by setting "readSubsFromDb": "false".

Metrics
-------
 Subscription Manager adds following statistic counters:

 Subscription create counters:
		- SubReqFromXapp: The total number of SubscriptionRequest messages received from xApp
		- SubRespToXapp: The total number of SubscriptionResponse messages sent to xApp
		- SubFailToXapp: The total number of SubscriptionFailure messages sent to xApp
		- SubReqToE2: The total number of SubscriptionRequest messages sent to E2Term
		- SubReReqToE2: The total number of SubscriptionRequest messages resent to E2Term
		- SubRespFromE2: The total number of SubscriptionResponse messages from E2Term
		- SubFailFromE2: The total number of SubscriptionFailure messages from E2Term
		- SubReqTimerExpiry: The total number of SubscriptionRequest timer expires
		- RouteCreateFail: The total number of subscription route create failure
		- RouteCreateUpdateFail: The total number of subscription route create update failure
		- MergedSubscriptions: The total number of merged Subscriptions

 Subscription delete counters:
		- SubDelReqFromXapp: The total number of SubscriptionDeleteResponse messages received from xApp
		- SubDelRespToXapp: The total number of SubscriptionDeleteResponse messages sent to xApp
		- SubDelReqToE2: The total number of SubscriptionDeleteRequest messages sent to E2Term
		- SubDelReReqToE2: The total number of SubscriptionDeleteRequest messages resent to E2Term
		- SubDelRespFromE2: The total number of SubscriptionDeleteResponse messages from E2Term
		- SubDelFailFromE2: The total number of SubscriptionDeleteFailure messages from E2Term
		- SubDelReqTimerExpiry: The total number of SubscriptionDeleteRequest timer expires
		- RouteDeleteFail: The total number of subscription route delete failure
		- RouteDeleteUpdateFail: The total number of subscription route delete update failure
		- UnmergedSubscriptions: The total number of unmerged Subscriptions

 SDL failure counters:
		- SDLWriteFailure: The total number of SDL write failures
		- SDLReadFailure: The total number of SDL read failures
		- SDLRemoveFailure: The total number of SDL read failures

Configurable parameters
-----------------------
 Subscription Manager has following configurable parameters.
   - Retry timeout for RIC Subscription Request message
      - e2tSubReqTimeout_ms: 2000 is the default value

   - Retry timeout for RIC Subscription Delete Request message
      - e2tSubDelReqTime_ms: 2000 is the default value

   - Waiting time for RIC Subscription Response and RIC Subscription Delete Response messages
      - e2tRecvMsgTimeout_ms: 2000 is the default value

   - Try count for RIC Subscription Request message   
      - e2tMaxSubReqTryCount: 2 is the default value

   - Try count for RIC Subscription Delete Request message   
      - e2tMaxSubDelReqTryCount: 2 is the default value
   
   - Are subscriptions read from database in Subscription Manager startup
      - readSubsFromDb: "true"  is the default value
 
 The parameters can be changed on the fly via Kubernetes Configmap. Default parameters values are defined in Helm chart

 Use following command to open Subscription Manager's Configmap in Nano editor. Firts change parameter and then store the
 change by pressing first Ctrl + o. Close editor by pressing the Ctrl + x. The change is visible in Subscription Manager's
 log after some 20 - 30 seconds.
 
 .. code-block:: none

  KUBE_EDITOR="nano" kubectl edit cm configmap-ricplt-submgr-submgrcfg -n ricplt

REST interface for debugging and testing
----------------------------------------
 Give following commands to get Subscription Manager pod's IP address

 .. code-block:: none

  kubectl get pods -A | grep submgr
  
  ricplt        submgr-75bccb84b6-n9vnt          1/1     Running             0          81m

  Syntax: kubectl exec -t -n ricplt <add-submgr-pod-name> -- cat /etc/hosts | grep submgr | awk '{print $1}'
  
  Example: kubectl exec -t -n ricplt submgr-75bccb84b6-n9vnt -- cat /etc/hosts | grep submgr | awk '{print $1}'

  10.244.0.181

 Get metrics

 .. code-block:: none

  Example: curl -s GET "http://10.244.0.181:8080/ric/v1/metrics"

 Get subscriptions

 .. code-block:: none

  Example: curl -X GET "http://10.244.0.181:8088/ric/v1/subscriptions"

 Delete single subscription from db

 .. code-block:: none

  Syntax: curl -X POST "http://10.244.0.181:8080/ric/v1/test/deletesubid={SubscriptionId}"
  
  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/deletesubid=1"

 Remove all subscriptions from db

 .. code-block:: none

  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/emptydb"

 Make Subscription Manager restart

 .. code-block:: none

  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/restart"

 Use this command to get Subscription Manager's log writings

 .. code-block:: none

   Example: kubectl logs -n ricplt submgr-75bccb84b6-n9vnt

 Logger level in configmap.yaml file in Helm chart is by default 2. It means that only info logs are printed.
 To see debug log writings it has to be changed to 4.

 .. code-block:: none

    "logger":
      "level": 4

RAN services explained
----------------------
  RIC hosted xApps may use the following RAN services from a RAN node:

  *  REPORT: RIC requests that RAN sends a REPORT message to RIC and continues further call processing in RAN after each occurrence of a defined SUBSCRIPTION
  *  INSERT: RIC requests that RAN sends an INSERT message to RIC and suspends further call processing in RAN after each occurrence of a defined SUBSCRIPTION
  *  CONTROL: RIC sends a Control message to RAN to initiate or resume call processing in RAN
  *  POLICY: RIC requests that RAN executes a specific POLICY during call processing in RAN after each occurrence of a defined SUBSCRIPTION

Supported E2 procedures and RAN services
----------------------------------------
    * RIC Subscription procedure with following RIC action types:

      - REPORT
      - POLICY
      - INSERT

    * RIC Subscription Delete procedure

    * Merge and delete of equal REPORT type subscriptions.

Recommendations for xApps
-------------------------

   * Recommended retry delay in xApp

     Recommended retry delay for xApp is >= 5 seconds. Length of supervising timers in Subscription Manager for the requests it sends to BTS is by default 2
     seconds. Subscription Manager makes one retry by default. There can be only one ongoing request towards per RAN other requests are queued in Subscription
     Manager.
