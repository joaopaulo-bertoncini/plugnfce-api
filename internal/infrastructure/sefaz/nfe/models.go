package nfe

import (
	"encoding/xml"
)

// NFCe represents the complete NFC-e structure
type NFCe struct {
	XMLName xml.Name `xml:"NFe"`
	InfNFe  InfNFe   `xml:"infNFe"`
}

// InfNFe represents the main NFC-e information block
type InfNFe struct {
	XMLName     xml.Name     `xml:"infNFe"`
	Versao      string       `xml:"versao,attr"`
	Id          string       `xml:"Id,attr"`
	Ide         Ide          `xml:"ide"`
	Emit        Emit         `xml:"emit"`
	Dest        *Dest        `xml:"dest,omitempty"`
	Det         []Det        `xml:"det"`
	Total       Total        `xml:"total"`
	Transp      Transp       `xml:"transp"`
	Cobr        *Cobr        `xml:"cobr,omitempty"`
	Pag         Pag          `xml:"pag"`
	InfIntermed *InfIntermed `xml:"infIntermed,omitempty"`
	InfRespTec  *InfRespTec  `xml:"infRespTec,omitempty"`
}

// Ide represents identification information
type Ide struct {
	CUF      string  `xml:"cUF"`
	CNF      string  `xml:"cNF"`
	NatOp    string  `xml:"natOp"`
	Mod      string  `xml:"mod"`
	Serie    string  `xml:"serie"`
	NNF      string  `xml:"nNF"`
	DhEmi    string  `xml:"dhEmi"`
	DhSaiEnt *string `xml:"dhSaiEnt,omitempty"`
	TpNF     string  `xml:"tpNF"`
	IdDest   string  `xml:"idDest"`
	CmunFG   string  `xml:"cMunFG"`
	TpImp    string  `xml:"tpImp"`
	TpEmis   string  `xml:"tpEmis"`
	Cdv      string  `xml:"cDV"`
	TpAmb    string  `xml:"tpAmb"`
	ProcEmi  string  `xml:"procEmi"`
	VerProc  string  `xml:"verProc"`
}

// Emit represents issuer information
type Emit struct {
	CNPJ      string    `xml:"CNPJ"`
	XNome     string    `xml:"xNome"`
	XFant     *string   `xml:"xFant,omitempty"`
	EnderEmit EnderEmit `xml:"enderEmit"`
	IE        string    `xml:"IE"`
	IM        *string   `xml:"IM,omitempty"`
	CNAE      *string   `xml:"CNAE,omitempty"`
	CRT       string    `xml:"CRT"`
}

// EnderEmit represents issuer address
type EnderEmit struct {
	XLgr    string  `xml:"xLgr"`
	Nro     string  `xml:"nro"`
	XCpl    *string `xml:"xCpl,omitempty"`
	XBairro string  `xml:"xBairro"`
	CMun    string  `xml:"cMun"`
	XMun    string  `xml:"xMun"`
	UF      string  `xml:"UF"`
	CEP     string  `xml:"CEP"`
	CPais   *string `xml:"cPais,omitempty"`
	XPais   *string `xml:"xPais,omitempty"`
	Fone    *string `xml:"fone,omitempty"`
}

// Dest represents destination information (optional for NFC-e)
type Dest struct {
	CNPJ      *string    `xml:"CNPJ,omitempty"`
	CPF       *string    `xml:"CPF,omitempty"`
	XNome     *string    `xml:"xNome,omitempty"`
	IndIEDest string     `xml:"indIEDest"`
	Email     *string    `xml:"email,omitempty"`
	EnderDest *EnderDest `xml:"enderDest,omitempty"`
}

// EnderDest represents destination address
type EnderDest struct {
	XLgr    string  `xml:"xLgr"`
	Nro     string  `xml:"nro"`
	XCpl    *string `xml:"xCpl,omitempty"`
	XBairro string  `xml:"xBairro"`
	CMun    string  `xml:"cMun"`
	XMun    string  `xml:"xMun"`
	UF      string  `xml:"UF"`
	CEP     string  `xml:"CEP"`
	CPais   *string `xml:"cPais,omitempty"`
	XPais   *string `xml:"xPais,omitempty"`
	Fone    *string `xml:"fone,omitempty"`
}

// Det represents detail/item information
type Det struct {
	NItem   string  `xml:"nItem,attr"`
	Prod    Prod    `xml:"prod"`
	Imposto Imposto `xml:"imposto"`
}

// Prod represents product information
type Prod struct {
	CProd    string  `xml:"cProd"`
	CEAN     *string `xml:"cEAN,omitempty"`
	XProd    string  `xml:"xProd"`
	NCM      string  `xml:"NCM"`
	CFOP     string  `xml:"CFOP"`
	UCom     string  `xml:"uCom"`
	QCom     string  `xml:"qCom"`
	VUnCom   string  `xml:"vUnCom"`
	VProd    string  `xml:"vProd"`
	CEANTrib *string `xml:"cEANTrib,omitempty"`
	UTrib    string  `xml:"uTrib"`
	QTrib    string  `xml:"qTrib"`
	VUnTrib  string  `xml:"vUnTrib"`
	IndTot   string  `xml:"indTot"`
	XPed     *string `xml:"xPed,omitempty"`
	NItemPed *string `xml:"nItemPed,omitempty"`
}

// Imposto represents tax information
type Imposto struct {
	VTotTrib *string `xml:"vTotTrib,omitempty"`
	ICMS     ICMS    `xml:"ICMS"`
	PIS      PIS     `xml:"PIS"`
	COFINS   COFINS  `xml:"COFINS"`
}

// ICMS represents ICMS tax
type ICMS struct {
	ICMS00    *ICMS00    `xml:"ICMS00,omitempty"`
	ICMS10    *ICMS10    `xml:"ICMS10,omitempty"`
	ICMS20    *ICMS20    `xml:"ICMS20,omitempty"`
	ICMS30    *ICMS30    `xml:"ICMS30,omitempty"`
	ICMS40    *ICMS40    `xml:"ICMS40,omitempty"`
	ICMS51    *ICMS51    `xml:"ICMS51,omitempty"`
	ICMS60    *ICMS60    `xml:"ICMS60,omitempty"`
	ICMS70    *ICMS70    `xml:"ICMS70,omitempty"`
	ICMS90    *ICMS90    `xml:"ICMS90,omitempty"`
	ICMSSN101 *ICMSSN101 `xml:"ICMSSN101,omitempty"`
	ICMSSN102 *ICMSSN102 `xml:"ICMSSN102,omitempty"`
	ICMSSN201 *ICMSSN201 `xml:"ICMSSN201,omitempty"`
	ICMSSN202 *ICMSSN202 `xml:"ICMSSN202,omitempty"`
	ICMSSN500 *ICMSSN500 `xml:"ICMSSN500,omitempty"`
	ICMSSN900 *ICMSSN900 `xml:"ICMSSN900,omitempty"`
}

// ICMS00 represents ICMS 00
type ICMS00 struct {
	Orig  string `xml:"orig"`
	CST   string `xml:"CST"`
	ModBC string `xml:"modBC"`
	VBC   string `xml:"vBC"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

// ICMS10 represents ICMS 10
type ICMS10 struct {
	Orig    string `xml:"orig"`
	CST     string `xml:"CST"`
	ModBC   string `xml:"modBC"`
	VBC     string `xml:"vBC"`
	PICMS   string `xml:"pICMS"`
	VICMS   string `xml:"vICMS"`
	ModBCST string `xml:"modBCST"`
	VBCST   string `xml:"vBCST"`
	PICMSST string `xml:"pICMSST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMS20 represents ICMS 20
type ICMS20 struct {
	Orig  string `xml:"orig"`
	CST   string `xml:"CST"`
	ModBC string `xml:"modBC"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

// ICMS30 represents ICMS 30
type ICMS30 struct {
	Orig    string `xml:"orig"`
	CST     string `xml:"CST"`
	ModBCST string `xml:"modBCST"`
	VBCST   string `xml:"vBCST"`
	PICMSST string `xml:"pICMSST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMS40 represents ICMS 40/41/50
type ICMS40 struct {
	Orig string `xml:"orig"`
	CST  string `xml:"CST"`
}

// ICMS51 represents ICMS 51
type ICMS51 struct {
	Orig  string `xml:"orig"`
	CST   string `xml:"CST"`
	ModBC string `xml:"modBC"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

// ICMS60 represents ICMS 60
type ICMS60 struct {
	Orig       string `xml:"orig"`
	CST        string `xml:"CST"`
	VBCSTRet   string `xml:"vBCSTRet"`
	VICMSSTRet string `xml:"vICMSSTRet"`
}

// ICMS70 represents ICMS 70
type ICMS70 struct {
	Orig    string `xml:"orig"`
	CST     string `xml:"CST"`
	ModBC   string `xml:"modBC"`
	VBC     string `xml:"vBC"`
	PICMS   string `xml:"pICMS"`
	VICMS   string `xml:"vICMS"`
	ModBCST string `xml:"modBCST"`
	PICMSST string `xml:"pICMSST"`
	VBCST   string `xml:"vBCST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMS90 represents ICMS 90
type ICMS90 struct {
	Orig    string `xml:"orig"`
	CST     string `xml:"CST"`
	ModBC   string `xml:"modBC"`
	VBC     string `xml:"vBC"`
	PICMS   string `xml:"pICMS"`
	VICMS   string `xml:"vICMS"`
	ModBCST string `xml:"modBCST"`
	PICMSST string `xml:"pICMSST"`
	VBCST   string `xml:"vBCST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMSSN101 represents ICMS SN 101
type ICMSSN101 struct {
	Orig  string `xml:"orig"`
	CSOSN string `xml:"CSOSN"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

// ICMSSN102 represents ICMS SN 102/103/300/400
type ICMSSN102 struct {
	Orig  string `xml:"orig"`
	CSOSN string `xml:"CSOSN"`
}

// ICMSSN201 represents ICMS SN 201
type ICMSSN201 struct {
	Orig    string `xml:"orig"`
	CSOSN   string `xml:"CSOSN"`
	ModBCST string `xml:"modBCST"`
	PICMSST string `xml:"pICMSST"`
	VBCST   string `xml:"vBCST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMSSN202 represents ICMS SN 202/203
type ICMSSN202 struct {
	Orig    string `xml:"orig"`
	CSOSN   string `xml:"CSOSN"`
	ModBCST string `xml:"modBCST"`
	PICMSST string `xml:"pICMSST"`
	VBCST   string `xml:"vBCST"`
	VICMSST string `xml:"vICMSST"`
}

// ICMSSN500 represents ICMS SN 500
type ICMSSN500 struct {
	Orig       string `xml:"orig"`
	CSOSN      string `xml:"CSOSN"`
	VBCSTRet   string `xml:"vBCSTRet"`
	VICMSSTRet string `xml:"vICMSSTRet"`
}

// ICMSSN900 represents ICMS SN 900
type ICMSSN900 struct {
	Orig    string `xml:"orig"`
	CSOSN   string `xml:"CSOSN"`
	ModBC   string `xml:"modBC"`
	VBC     string `xml:"vBC"`
	PICMS   string `xml:"pICMS"`
	VICMS   string `xml:"vICMS"`
	ModBCST string `xml:"modBCST"`
	PICMSST string `xml:"pICMSST"`
	VBCST   string `xml:"vBCST"`
	VICMSST string `xml:"vICMSST"`
}

// PIS represents PIS tax
type PIS struct {
	PISAliq *PISAliq `xml:"PISAliq,omitempty"`
	PISQtde *PISQtde `xml:"PISQtde,omitempty"`
	PISNT   *PISNT   `xml:"PISNT,omitempty"`
	PISOutr *PISOutr `xml:"PISOutr,omitempty"`
}

// PISAliq represents PIS by aliquot
type PISAliq struct {
	CST  string `xml:"CST"`
	VBC  string `xml:"vBC"`
	PPIS string `xml:"pPIS"`
	VPIS string `xml:"vPIS"`
}

// PISQtde represents PIS by quantity
type PISQtde struct {
	CST       string `xml:"CST"`
	QBCProd   string `xml:"qBCProd"`
	VAliqProd string `xml:"vAliqProd"`
	VPIS      string `xml:"vPIS"`
}

// PISNT represents PIS not taxable
type PISNT struct {
	CST string `xml:"CST"`
}

// PISOutr represents PIS other
type PISOutr struct {
	CST  string `xml:"CST"`
	VBC  string `xml:"vBC"`
	PPIS string `xml:"pPIS"`
	VPIS string `xml:"vPIS"`
}

// COFINS represents COFINS tax
type COFINS struct {
	COFINSAliq *COFINSAliq `xml:"COFINSAliq,omitempty"`
	COFINSQtde *COFINSQtde `xml:"COFINSQtde,omitempty"`
	COFINSNT   *COFINSNT   `xml:"COFINSNT,omitempty"`
	COFINSOutr *COFINSOutr `xml:"COFINSOutr,omitempty"`
}

// COFINSAliq represents COFINS by aliquot
type COFINSAliq struct {
	CST     string `xml:"CST"`
	VBC     string `xml:"vBC"`
	PCOFINS string `xml:"pCOFINS"`
	VCOFINS string `xml:"vCOFINS"`
}

// COFINSQtde represents COFINS by quantity
type COFINSQtde struct {
	CST       string `xml:"CST"`
	QBCProd   string `xml:"qBCProd"`
	VAliqProd string `xml:"vAliqProd"`
	VCOFINS   string `xml:"vCOFINS"`
}

// COFINSNT represents COFINS not taxable
type COFINSNT struct {
	CST string `xml:"CST"`
}

// COFINSOutr represents COFINS other
type COFINSOutr struct {
	CST     string `xml:"CST"`
	VBC     string `xml:"vBC"`
	PCOFINS string `xml:"pCOFINS"`
	VCOFINS string `xml:"vCOFINS"`
}

// Total represents total information
type Total struct {
	ICMSTot ICMSTot `xml:"ICMSTot"`
}

// ICMSTot represents ICMS total
type ICMSTot struct {
	VBC          string  `xml:"vBC"`
	VICMS        string  `xml:"vICMS"`
	VICMSDescont *string `xml:"vICMSDescont,omitempty"`
	VBCST        string  `xml:"vBCST"`
	VST          string  `xml:"vST"`
	VProd        string  `xml:"vProd"`
	VFrete       *string `xml:"vFrete,omitempty"`
	VSeg         *string `xml:"vSeg,omitempty"`
	VDesc        *string `xml:"vDesc,omitempty"`
	VII          *string `xml:"vII,omitempty"`
	VIPI         *string `xml:"vIPI,omitempty"`
	VIPIDevol    *string `xml:"vIPIDevol,omitempty"`
	VPIS         string  `xml:"vPIS"`
	VCOFINS      string  `xml:"vCOFINS"`
	VOutro       *string `xml:"vOutro,omitempty"`
	VNF          string  `xml:"vNF"`
	VTotTrib     *string `xml:"vTotTrib,omitempty"`
}

// Transp represents transport information
type Transp struct {
	ModFrete string `xml:"modFrete"`
}

// Cobr represents billing information (optional)
type Cobr struct {
	Fat *Fat  `xml:"fat,omitempty"`
	Dup []Dup `xml:"dup,omitempty"`
}

// Fat represents billing header
type Fat struct {
	NFat  string  `xml:"nFat"`
	VOrig string  `xml:"vOrig"`
	VDesc *string `xml:"vDesc,omitempty"`
	VLiq  string  `xml:"vLiq"`
}

// Dup represents billing installment
type Dup struct {
	NDup  string `xml:"nDup"`
	DVenc string `xml:"dVenc"`
	VDup  string `xml:"vDup"`
}

// Pag represents payment information
type Pag struct {
	DetPag []DetPag `xml:"detPag"`
	VTroco *string  `xml:"vTroco,omitempty"`
}

// DetPag represents payment detail
type DetPag struct {
	TPag string `xml:"tPag"`
	VPag string `xml:"vPag"`
	Card *Card  `xml:"card,omitempty"`
}

// Card represents card payment information
type Card struct {
	TpIntegra string  `xml:"tpIntegra"`
	CNPJ      *string `xml:"CNPJ,omitempty"`
	TBand     *string `xml:"tBand,omitempty"`
	CAut      *string `xml:"cAut,omitempty"`
}

// InfIntermed represents intermediary information
type InfIntermed struct {
	CNPJ         string  `xml:"CNPJ"`
	XNome        string  `xml:"xNome"`
	IdCadIntTran *string `xml:"idCadIntTran,omitempty"`
}

// InfRespTec represents technical responsible information
type InfRespTec struct {
	CNPJ     string `xml:"CNPJ"`
	XContato string `xml:"xContato"`
	Email    string `xml:"email"`
	Fone     string `xml:"fone"`
}

// NFCeInput represents the input data for NFC-e generation
type NFCeInput struct {
	UF           string
	Ambiente     string
	Emitente     EmitenteInput
	Destinatario *DestinatarioInput
	Itens        []ItemInput
	Pagamentos   []PagamentoInput
	Transp       TranspInput
	InfIntermed  *InfIntermedInput
	InfRespTec   *InfRespTecInput
}

// EmitenteInput represents issuer input data
type EmitenteInput struct {
	CNPJ      string
	XNome     string
	XFant     *string
	EnderEmit EnderEmitInput
	IE        string
	IM        *string
	CNAE      *string
	CRT       string
}

// EnderEmitInput represents issuer address input
type EnderEmitInput struct {
	XLgr    string
	Nro     string
	XCpl    *string
	XBairro string
	CMun    string
	XMun    string
	UF      string
	CEP     string
	CPais   *string
	XPais   *string
	Fone    *string
}

// DestinatarioInput represents destination input data
type DestinatarioInput struct {
	CNPJ      *string
	CPF       *string
	XNome     *string
	IndIEDest string
	Email     *string
	EnderDest *EnderDestInput
}

// EnderDestInput represents destination address input
type EnderDestInput struct {
	XLgr    string
	Nro     string
	XCpl    *string
	XBairro string
	CMun    string
	XMun    string
	UF      string
	CEP     string
	CPais   *string
	XPais   *string
	Fone    *string
}

// ItemInput represents item input data
type ItemInput struct {
	CProd    string
	CEAN     *string
	XProd    string
	NCM      string
	CFOP     string
	UCom     string
	QCom     string
	VUnCom   string
	VProd    string
	CEANTrib *string
	UTrib    string
	QTrib    string
	VUnTrib  string
	IndTot   string
	XPed     *string
	NItemPed *string
	Imposto  ImpostoInput
}

// ImpostoInput represents tax input data
type ImpostoInput struct {
	VTotTrib *string
	ICMS     ICMSInput
	PIS      PISInput
	COFINS   COFINSInput
}

// ICMSInput represents ICMS input
type ICMSInput struct {
	Tipo    string // "ICMS00", "ICMS10", etc.
	Orig    string
	CST     string
	ModBC   *string
	VBC     *string
	PICMS   *string
	VICMS   *string
	ModBCST *string
	VBCST   *string
	PICMSST *string
	VICMSST *string
}

// PISInput represents PIS input
type PISInput struct {
	Tipo      string // "PISAliq", "PISQtde", "PISNT", "PISOutr"
	CST       string
	VBC       *string
	PPIS      *string
	VPIS      *string
	QBCProd   *string
	VAliqProd *string
}

// COFINSInput represents COFINS input
type COFINSInput struct {
	Tipo      string // "COFINSAliq", "COFINSQtde", "COFINSNT", "COFINSOutr"
	CST       string
	VBC       *string
	PCOFINS   *string
	VCOFINS   *string
	QBCProd   *string
	VAliqProd *string
}

// PagamentoInput represents payment input data
type PagamentoInput struct {
	TPag string
	VPag string
	Card *CardInput
}

// CardInput represents card payment input
type CardInput struct {
	TpIntegra string
	CNPJ      *string
	TBand     *string
	CAut      *string
}

// TranspInput represents transport input
type TranspInput struct {
	ModFrete string
}

// InfIntermedInput represents intermediary input
type InfIntermedInput struct {
	CNPJ         string
	XNome        string
	IdCadIntTran *string
}

// InfRespTecInput represents technical responsible input
type InfRespTecInput struct {
	CNPJ     string
	XContato string
	Email    string
	Fone     string
}
