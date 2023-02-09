package search

type SearchResults struct {
	AccountActivities []AccountActivity
	AccessProfiles    []AccessProfile
	Entitlements      []Entitlement
	Events            []Event
	Identities        []Identity
	Roles             []Role
}

type AccessProfile struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	Type                    string `json:"_type"`
	Description             string `json:"description"`
	Created                 string `json:"created"`
	Modified                string `json:"modified"`
	Synced                  string `json:"synced"`
	Enabled                 bool   `json:"enabled"`
	Requestable             bool   `json:"requestable"`
	RequestCommentsRequired bool   `json:"requestCommentsRequired"`
	Owner                   struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Email string `json:"email"`
	} `json:"owner"`
	Source struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Entitlements []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Attribute   string `json:"attribute"`
		Value       string `json:"value"`
	} `json:"entitlements"`
	EntitlementCount int      `json:"entitlementCount"`
	Tags             []string `json:"tags"`
}

type AccountActivity struct {
	Requester struct {
		Name string `json:"name,omitempty"`
		ID   string `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"requester,omitempty"`
	Sources         string   `json:"sources,omitempty"`
	Created         string   `json:"created,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	AccountRequests []struct {
		Result struct {
			Status string `json:"status,omitempty"`
		} `json:"result,omitempty"`
		AccountID         string `json:"accountId,omitempty"`
		Op                string `json:"op,omitempty"`
		AttributeRequests []struct {
			Op    string `json:"op,omitempty"`
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"attributeRequests,omitempty"`
		ProvisioningTarget struct {
			Name string `json:"name,omitempty"`
			ID   string `json:"id,omitempty"`
			Type string `json:"type,omitempty"`
		} `json:"provisioningTarget,omitempty"`
		Source struct {
			Name string `json:"name,omitempty"`
			ID   string `json:"id,omitempty"`
			Type string `json:"type,omitempty"`
		} `json:"source,omitempty"`
	} `json:"accountRequests,omitempty"`
	Stage            string `json:"stage,omitempty"`
	OriginalRequests []struct {
		Result struct {
			Status string `json:"status,omitempty"`
		} `json:"result,omitempty"`
		AccountID string `json:"accountId,omitempty"`
		Op        string `json:"op,omitempty"`
		Source    struct {
			Name string `json:"name,omitempty"`
			ID   string `json:"id,omitempty"`
		} `json:"source,omitempty"`
		AttributeRequests []struct {
			Op    string `json:"op,omitempty"`
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"attributeRequests,omitempty"`
	} `json:"originalRequests,omitempty"`
	ExpansionItems []interface{} `json:"expansionItems,omitempty"`
	Approvals      []struct {
		AttributeRequest struct {
			Op    string `json:"op,omitempty"`
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"attributeRequest,omitempty"`
		Source struct {
			Name string `json:"name,omitempty"`
			ID   string `json:"id,omitempty"`
		} `json:"source,omitempty"`
	} `json:"approvals,omitempty"`
	Recipient struct {
		Name string `json:"name,omitempty"`
		ID   string `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"recipient,omitempty"`
	Action         string `json:"action,omitempty"`
	Modified       string `json:"modified,omitempty"`
	ID             string `json:"id,omitempty"`
	TrackingNumber string `json:"trackingNumber,omitempty"`
	Status         string `json:"status,omitempty"`
	Pod            string `json:"pod,omitempty"`
	Org            string `json:"org,omitempty"`
	Synced         string `json:"synced,omitempty"`
	Type           string `json:"_type,omitempty"`
	Type0          string `json:"type,omitempty"`
	Version        string `json:"_version,omitempty"`
}

type Entitlement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"_type"`
	Description string `json:"description"`
	Attribute   string `json:"attribute"`
	Value       string `json:"value"`
	Modified    string `json:"modified"`
	Synced      string `json:"synced"`
	DisplayName string `json:"displayName"`
	Source      struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Privileged    bool     `json:"privileged"`
	IdentityCount int      `json:"identityCount"`
	Tags          []string `json:"tags"`
}

type Event struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"_type"`
	Created string `json:"created"`
	Synced  string `json:"synced"`
	Action  string `json:"action"`
	Type0   string `json:"type"`
	Actor   struct {
		Name string `json:"name"`
	} `json:"actor"`
	Target struct {
		Name string `json:"name"`
	} `json:"target"`
	Stack          string `json:"stack"`
	TrackingNumber string `json:"trackingNumber"`
	IPAddress      string `json:"ipAddress"`
	Details        string `json:"details"`
	Attributes     struct {
		SourceName string `json:"sourceName"`
	} `json:"attributes"`
	Objects       []string `json:"objects"`
	Operation     string   `json:"operation"`
	Status        string   `json:"status"`
	TechnicalName string   `json:"technicalName"`
}

type Identity struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Type            string      `json:"_type"`
	FirstName       string      `json:"firstName"`
	LastName        string      `json:"lastName"`
	DisplayName     string      `json:"displayName"`
	Email           string      `json:"email"`
	Created         string      `json:"created"`
	Modified        string      `json:"modified"`
	Synced          string      `json:"synced"`
	Phone           string      `json:"phone"`
	Inactive        bool        `json:"inactive"`
	Protected       bool        `json:"protected"`
	Status          string      `json:"status"`
	EmployeeNumber  string      `json:"employeeNumber"`
	Manager         interface{} `json:"manager"`
	IsManager       bool        `json:"isManager"`
	IdentityProfile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"identityProfile"`
	Source struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Attributes struct {
		UID                      string      `json:"uid"`
		Firstname                string      `json:"firstname"`
		CloudAuthoritativeSource string      `json:"cloudAuthoritativeSource"`
		CloudStatus              string      `json:"cloudStatus"`
		IplanetAmUserAliasList   interface{} `json:"iplanet-am-user-alias-list"`
		DisplayName              string      `json:"displayName"`
		InternalCloudStatus      string      `json:"internalCloudStatus"`
		WorkPhone                string      `json:"workPhone"`
		Email                    string      `json:"email"`
		Lastname                 string      `json:"lastname"`
	} `json:"attributes"`
	ProcessingState   interface{} `json:"processingState"`
	ProcessingDetails interface{} `json:"processingDetails"`
	Accounts          []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		AccountID string `json:"accountId"`
		Source    struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"source"`
		Disabled              bool   `json:"disabled"`
		Locked                bool   `json:"locked"`
		Privileged            bool   `json:"privileged"`
		ManuallyCorrelated    bool   `json:"manuallyCorrelated"`
		PasswordLastSet       string `json:"passwordLastSet"`
		EntitlementAttributes struct {
			MemberOf []string `json:"memberOf"`
		} `json:"entitlementAttributes"`
		Created string `json:"created"`
	} `json:"accounts"`
	AccountCount int `json:"accountCount"`
	Apps         []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Source struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source"`
		Account struct {
			ID        string `json:"id"`
			AccountID string `json:"accountId"`
		} `json:"account"`
	} `json:"apps"`
	AppCount int `json:"appCount"`
	Access   []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Source      struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source,omitempty"`
		Owner struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"owner,omitempty"`
		Privileged bool   `json:"privileged,omitempty"`
		Attribute  string `json:"attribute,omitempty"`
		Value      string `json:"value,omitempty"`
		Standalone bool   `json:"standalone,omitempty"`
		Disabled   bool   `json:"disabled,omitempty"`
	} `json:"access"`
	AccessCount        int      `json:"accessCount"`
	AccessProfileCount int      `json:"accessProfileCount"`
	EntitlementCount   int      `json:"entitlementCount"`
	RoleCount          int      `json:"roleCount"`
	Tags               []string `json:"tags"`
}

type Role struct {
	ID                      string      `json:"id"`
	Name                    string      `json:"name"`
	Type                    string      `json:"_type"`
	Description             string      `json:"description"`
	Created                 string      `json:"created"`
	Modified                interface{} `json:"modified"`
	Synced                  string      `json:"synced"`
	Enabled                 bool        `json:"enabled"`
	Requestable             bool        `json:"requestable"`
	RequestCommentsRequired bool        `json:"requestCommentsRequired"`
	Owner                   struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Email string `json:"email"`
	} `json:"owner"`
	AccessProfiles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"accessProfiles"`
	AccessProfileCount int      `json:"accessProfileCount"`
	Tags               []string `json:"tags"`
}
