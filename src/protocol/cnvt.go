package protocol

type SimpleConverter int

// HTTPToCustom : TODO implement.
func (cnvt SimpleConverter) HTTPToCustom(httpreq HttpPacks) (req CustRequest) {

}

// CustomToDubbo : TODO implement.
func (cnvt SimpleConverter) CustomToDubbo(custreq CustRequest) (dubboreq DubboPacks) {

}

// DubboToCustom : TODO implement.
func (cnvt SimpleConverter) DubboToCustom(dubboresp DubboPacks) (custresp CustResponse) {

}

// CustomToHTTP : TODO implement.
func (cnvt SimpleConverter) CustomToHTTP(resp CustResponse) (httpresp HttpPacks) {

}
