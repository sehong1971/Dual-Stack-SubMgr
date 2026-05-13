package control

import (
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

const (
	cSubReqFromXapp         string = "SubReqFromXapp"
	cRestSubReqFromXapp     string = "RestSubReqFromXapp"
	cSubFailToXapp          string = "SubFailToXapp"
	cSubRespToXapp          string = "SubRespToXapp"
	cRestSubRespToXapp      string = "RestSubRespToXapp"
	cRestSubFailToXapp      string = "RestSubFailToXapp"
	cRestReqRejDueE2Down    string = "RestReqRejDueE2Down"
	cRestSubNotifToXapp     string = "RestSubNotifToXapp"
	cRestSubFailNotifToXapp string = "RestSubFailNotifToXapp"
	cSubReqToE2             string = "SubReqToE2"
	cSubReReqToE2           string = "SubReReqToE2"
	cSubRespFromE2          string = "SubRespFromE2"
	cPartialSubRespFromE2   string = "PartialSubRespFromE2"
	cSubFailFromE2          string = "SubFailFromE2"
	cSubReqTimerExpiry      string = "SubReqTimerExpiry"
	cRouteCreateFail        string = "RouteCreateFail"
	cRouteCreateUpdateFail  string = "RouteCreateUpdateFail"
	cMergedSubscriptions    string = "MergedSubscriptions"
	cDuplicateE2SubReq      string = "DuplicateE2SubReq"
	cSubDelReqFromXapp      string = "SubDelReqFromXapp"
	cSubDelRespToXapp       string = "SubDelRespToXapp"
	cRestSubDelReqFromXapp  string = "RestSubDelReqFromXapp"
	cRestSubDelRespToXapp   string = "RestSubDelRespToXapp"
	cRestSubDelFailToXapp   string = "RestSubDelFailToXapp"
	cSubDelReqToE2          string = "SubDelReqToE2"
	cSubDelReReqToE2        string = "SubDelReReqToE2"
	cSubDelRespFromE2       string = "SubDelRespFromE2"
	cSubDelFailFromE2       string = "SubDelFailFromE2"
	cSubDelReqTimerExpiry   string = "SubDelReqTimerExpiry"
	cSubDelRequFromE2       string = "SubDelRequiredFromE2"
	cRouteDeleteFail        string = "RouteDeleteFail"
	cRouteDeleteUpdateFail  string = "RouteDeleteUpdateFail"
	cUnmergedSubscriptions  string = "UnmergedSubscriptions"
	cSDLWriteFailure        string = "SDLWriteFailure"
	cSDLReadFailure         string = "SDLReadFailure"
	cSDLRemoveFailure       string = "SDLRemoveFailure"
	cE2StateChangedToUp     string = "E2StateChangedToUp"
	cE2StateChangedToDown   string = "E2StateChangedToDown"
	cE2StateUnderReset      string = "E2StateChangedToUnderReset"
)

func GetMetricsOpts() []xapp.CounterOpts {
	return []xapp.CounterOpts{

		// Subscrition create counters
		{Name: cSubReqFromXapp, Help: "The total number of SubscriptionRequest messages received from xApp"},
		{Name: cSubRespToXapp, Help: "The total number of SubscriptionResponse messages sent to xApp"},
		{Name: cSubFailToXapp, Help: "The total number of SubscriptionFailure messages sent to xApp"},
		{Name: cRestSubReqFromXapp, Help: "The total number of Rest SubscriptionRequest messages received from xApp"},
		{Name: cRestSubRespToXapp, Help: "The total number of Rest SubscriptionResponse messages sent to xApp"},
		{Name: cRestSubFailToXapp, Help: "The total number of Rest SubscriptionFailure messages sent to xApp"},
		{Name: cRestReqRejDueE2Down, Help: "The total number of Rest SubscriptionRequest messages rejected due E2 Interface down"},
		{Name: cRestSubNotifToXapp, Help: "The total number of successful Rest SubscriptionNotification messages sent to xApp"},
		{Name: cRestSubFailNotifToXapp, Help: "The total number of failure Rest SubscriptionNotification messages sent to xApp"},
		{Name: cSubReqToE2, Help: "The total number of SubscriptionRequest messages sent to E2Term"},
		{Name: cSubReReqToE2, Help: "The total number of SubscriptionRequest messages resent to E2Term"},
		{Name: cPartialSubRespFromE2, Help: "The total number of partial SubscriptionResponse messages from E2Term"},
		{Name: cSubRespFromE2, Help: "The total number of SubscriptionResponse messages from E2Term"},
		{Name: cSubFailFromE2, Help: "The total number of SubscriptionFailure messages from E2Term"},
		{Name: cSubReqTimerExpiry, Help: "The total number of SubscriptionRequest timer expires"},
		{Name: cRouteCreateFail, Help: "The total number of subscription route create failure"},
		{Name: cRouteCreateUpdateFail, Help: "The total number of subscription route create update failure"},
		{Name: cMergedSubscriptions, Help: "The total number of merged Subscriptions"},
		{Name: cDuplicateE2SubReq, Help: "The total number of same E2 SubscriptionRequest messages from same xApp"},

		// Subscrition delete counters
		{Name: cSubDelReqFromXapp, Help: "The total number of SubscriptionDeleteRequest messages received from xApp"},
		{Name: cSubDelRespToXapp, Help: "The total number of SubscriptionDeleteResponse messages sent to xApp"},
		{Name: cRestSubDelReqFromXapp, Help: "The total number of Rest SubscriptionDeleteRequest messages received from xApp"},
		{Name: cRestSubDelRespToXapp, Help: "The total number of Rest SubscriptionDeleteResponse messages sent to xApp"},
		{Name: cRestSubDelFailToXapp, Help: "The total number of Rest SubscriptionDeleteFailure messages sent to xApp"},
		{Name: cSubDelReqToE2, Help: "The total number of SubscriptionDeleteRequest messages sent to E2Term"},
		{Name: cSubDelReReqToE2, Help: "The total number of SubscriptionDeleteRequest messages resent to E2Term"},
		{Name: cSubDelRespFromE2, Help: "The total number of SubscriptionDeleteResponse messages from E2Term"},
		{Name: cSubDelFailFromE2, Help: "The total number of SubscriptionDeleteFailure messages from E2Term"},
		{Name: cSubDelReqTimerExpiry, Help: "The total number of SubscriptionDeleteRequest timer expires"},
		{Name: cSubDelRequFromE2, Help: "The total number of SubscriptionDeleteRequired messages from E2Term"},
		{Name: cRouteDeleteFail, Help: "The total number of subscription route delete failure"},
		{Name: cRouteDeleteUpdateFail, Help: "The total number of subscription route delete update failure"},
		{Name: cUnmergedSubscriptions, Help: "The total number of unmerged Subscriptions"},

		// SDL failure counters
		{Name: cSDLWriteFailure, Help: "The total number of SDL write failures"},
		{Name: cSDLReadFailure, Help: "The total number of SDL read failures"},
		{Name: cSDLRemoveFailure, Help: "The total number of SDL remove failures"},

		// E2 interface state counters
		{Name: cE2StateChangedToUp, Help: "The total number of E2 interface change connected state"},
		{Name: cE2StateChangedToDown, Help: "The total number of E2 interface change disconnected state"},
		{Name: cE2StateUnderReset, Help: "The total number of E2 interface change under reset state"},
	}
}

func (c *Control) UpdateCounter(counterName string) {
	xapp.Logger.Debug("Add counterName=%v", counterName)
	c.Counters[counterName].Inc()
}
