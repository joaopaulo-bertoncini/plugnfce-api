package nfe

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// Builder handles NFC-e XML construction
type Builder interface {
	BuildNFCe(input NFCeInput, companyID string) (*NFCe, error)
	GenerateChaveAcesso(uf, cnpj, serie, nNF, tpEmis, cNF string, dhEmi time.Time) (string, error)
	CalculateDV(chave string) string
}

// builder implements Builder interface
type builder struct {
	companyRepo ports.CompanyRepository
}

// NewBuilder creates a new NFC-e builder
func NewBuilder(companyRepo ports.CompanyRepository) Builder {
	return &builder{
		companyRepo: companyRepo,
	}
}

// BuildNFCe builds a complete NFC-e XML from input data
func (b *builder) BuildNFCe(input NFCeInput, companyID string) (*NFCe, error) {
	// Get next sequential number (NNF) from database
	nextNumber, err := b.companyRepo.GetNextNFCeNumber(context.Background(), companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next NFC-e number: %w", err)
	}
	nNF := strconv.FormatInt(nextNumber, 10)

	// Generate random number for CNF (8 digits)
	cNF := b.generateCNF()

	// Generate chave de acesso
	chave, err := b.GenerateChaveAcesso(
		input.UF,
		input.Emitente.CNPJ,
		"1", // serie - should be configurable
		nNF,
		"1", // tpEmis - normal
		cNF,
		time.Now(), // dhEmi
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chave acesso: %w", err)
	}

	// Build NFC-e structure
	nfce := &NFCe{
		InfNFe: InfNFe{
			Versao: "4.00",
			Id:     "NFe" + chave,
			Ide:    b.buildIde(input, nNF, cNF, chave),
			Emit:   b.buildEmit(input.Emitente),
			Det:    b.buildDet(input.Itens),
			Total:  b.buildTotal(input.Itens),
			Transp: b.buildTransp(input.Transp),
			Pag:    b.buildPag(input.Pagamentos),
		},
	}

	// Add optional fields
	if input.Destinatario != nil {
		dest := b.buildDest(*input.Destinatario)
		nfce.InfNFe.Dest = &dest
	}

	if input.InfIntermed != nil {
		infIntermed := b.buildInfIntermed(*input.InfIntermed)
		nfce.InfNFe.InfIntermed = &infIntermed
	}

	if input.InfRespTec != nil {
		infRespTec := b.buildInfRespTec(*input.InfRespTec)
		nfce.InfNFe.InfRespTec = &infRespTec
	}

	return nfce, nil
}

// buildIde builds identification block
func (b *builder) buildIde(input NFCeInput, nNF, cNF, chave string) Ide {
	// Get municipality code based on UF
	cMunFG := b.getMunicipioFG(input.UF)

	// Determine emission type based on contingency
	tpEmis := "1" // Normal emission
	if input.Contingency {
		if input.ContingencyType == "SVC-AN" {
			tpEmis = "6" // SVC-AN contingency
		} else if input.ContingencyType == "SVC-RS" {
			tpEmis = "7" // SVC-RS contingency
		}
	}

	return Ide{
		CUF:     b.getCUF(input.UF),
		CNF:     cNF,
		NatOp:   "VENDA",
		Mod:     "65", // NFC-e
		Serie:   "1",
		NNF:     nNF,
		DhEmi:   time.Now().Format(time.RFC3339),
		TpNF:    "1", // Saída
		IdDest:  "1", // Interna
		CmunFG:  cMunFG,
		TpImp:   "4",                       // DANFE NFC-e
		TpEmis:  tpEmis,                    // Normal or contingency
		Cdv:     b.CalculateDV(chave[:43]), // Last digit of chave
		TpAmb:   input.Ambiente,
		ProcEmi: "0", // Emissão própria
		VerProc: "1.0.0",
	}
}

// buildEmit builds issuer block
func (b *builder) buildEmit(emit EmitenteInput) Emit {
	return Emit{
		CNPJ:  emit.CNPJ,
		XNome: emit.XNome,
		XFant: emit.XFant,
		EnderEmit: EnderEmit{
			XLgr:    emit.EnderEmit.XLgr,
			Nro:     emit.EnderEmit.Nro,
			XCpl:    emit.EnderEmit.XCpl,
			XBairro: emit.EnderEmit.XBairro,
			CMun:    emit.EnderEmit.CMun,
			XMun:    emit.EnderEmit.XMun,
			UF:      emit.EnderEmit.UF,
			CEP:     emit.EnderEmit.CEP,
			CPais:   emit.EnderEmit.CPais,
			XPais:   emit.EnderEmit.XPais,
			Fone:    emit.EnderEmit.Fone,
		},
		IE:   emit.IE,
		IM:   emit.IM,
		CNAE: emit.CNAE,
		CRT:  emit.CRT,
	}
}

// buildDest builds destination block
func (b *builder) buildDest(dest DestinatarioInput) Dest {
	return Dest{
		CNPJ:      dest.CNPJ,
		CPF:       dest.CPF,
		XNome:     dest.XNome,
		IndIEDest: dest.IndIEDest,
		Email:     dest.Email,
		EnderDest: func() *EnderDest {
			if dest.EnderDest == nil {
				return nil
			}
			return &EnderDest{
				XLgr:    dest.EnderDest.XLgr,
				Nro:     dest.EnderDest.Nro,
				XCpl:    dest.EnderDest.XCpl,
				XBairro: dest.EnderDest.XBairro,
				CMun:    dest.EnderDest.CMun,
				XMun:    dest.EnderDest.XMun,
				UF:      dest.EnderDest.UF,
				CEP:     dest.EnderDest.CEP,
				CPais:   dest.EnderDest.CPais,
				XPais:   dest.EnderDest.XPais,
				Fone:    dest.EnderDest.Fone,
			}
		}(),
	}
}

// buildDet builds detail/items block
func (b *builder) buildDet(itens []ItemInput) []Det {
	det := make([]Det, len(itens))
	for i, item := range itens {
		det[i] = Det{
			NItem: strconv.Itoa(i + 1),
			Prod: Prod{
				CProd:    item.CProd,
				CEAN:     item.CEAN,
				XProd:    item.XProd,
				NCM:      item.NCM,
				CFOP:     item.CFOP,
				UCom:     item.UCom,
				QCom:     item.QCom,
				VUnCom:   item.VUnCom,
				VProd:    item.VProd,
				CEANTrib: item.CEANTrib,
				UTrib:    item.UTrib,
				QTrib:    item.QTrib,
				VUnTrib:  item.VUnTrib,
				IndTot:   item.IndTot,
				XPed:     item.XPed,
				NItemPed: item.NItemPed,
			},
			Imposto: b.buildImposto(item.Imposto),
		}
	}
	return det
}

// buildImposto builds tax block
func (b *builder) buildImposto(imposto ImpostoInput) Imposto {
	imp := Imposto{
		VTotTrib: imposto.VTotTrib,
	}

	// Build ICMS
	imp.ICMS = b.buildICMS(imposto.ICMS)

	// Build PIS
	imp.PIS = b.buildPIS(imposto.PIS)

	// Build COFINS
	imp.COFINS = b.buildCOFINS(imposto.COFINS)

	return imp
}

// buildICMS builds ICMS tax block
func (b *builder) buildICMS(icms ICMSInput) ICMS {
	var result ICMS

	switch icms.Tipo {
	case "ICMS00":
		result.ICMS00 = &ICMS00{
			Orig:  icms.Orig,
			CST:   icms.CST,
			ModBC: *icms.ModBC,
			VBC:   *icms.VBC,
			PICMS: *icms.PICMS,
			VICMS: *icms.VICMS,
		}
	case "ICMS10":
		result.ICMS10 = &ICMS10{
			Orig:    icms.Orig,
			CST:     icms.CST,
			ModBC:   *icms.ModBC,
			VBC:     *icms.VBC,
			PICMS:   *icms.PICMS,
			VICMS:   *icms.VICMS,
			ModBCST: *icms.ModBCST,
			VBCST:   *icms.VBCST,
			PICMSST: *icms.PICMSST,
			VICMSST: *icms.VICMSST,
		}
	case "ICMS20":
		result.ICMS20 = &ICMS20{
			Orig:  icms.Orig,
			CST:   icms.CST,
			ModBC: *icms.ModBC,
			PICMS: *icms.PICMS,
			VICMS: *icms.VICMS,
		}
	case "ICMS40", "ICMS41", "ICMS50":
		result.ICMS40 = &ICMS40{
			Orig: icms.Orig,
			CST:  icms.CST,
		}
	case "ICMS51":
		result.ICMS51 = &ICMS51{
			Orig:  icms.Orig,
			CST:   icms.CST,
			ModBC: *icms.ModBC,
			PICMS: *icms.PICMS,
			VICMS: *icms.VICMS,
		}
	case "ICMS60":
		result.ICMS60 = &ICMS60{
			Orig:       icms.Orig,
			CST:        icms.CST,
			VBCSTRet:   *icms.VBCST,
			VICMSSTRet: *icms.VICMSST,
		}
	case "ICMS70":
		result.ICMS70 = &ICMS70{
			Orig:    icms.Orig,
			CST:     icms.CST,
			ModBC:   *icms.ModBC,
			VBC:     *icms.VBC,
			PICMS:   *icms.PICMS,
			VICMS:   *icms.VICMS,
			ModBCST: *icms.ModBCST,
			PICMSST: *icms.PICMSST,
			VBCST:   *icms.VBCST,
			VICMSST: *icms.VICMSST,
		}
	case "ICMS90":
		result.ICMS90 = &ICMS90{
			Orig:    icms.Orig,
			CST:     icms.CST,
			ModBC:   *icms.ModBC,
			VBC:     *icms.VBC,
			PICMS:   *icms.PICMS,
			VICMS:   *icms.VICMS,
			ModBCST: *icms.ModBCST,
			PICMSST: *icms.PICMSST,
			VBCST:   *icms.VBCST,
			VICMSST: *icms.VICMSST,
		}
	case "ICMSSN101":
		result.ICMSSN101 = &ICMSSN101{
			Orig:  icms.Orig,
			CSOSN: icms.CST,
			PICMS: *icms.PICMS,
			VICMS: *icms.VICMS,
		}
	case "ICMSSN102", "ICMSSN103", "ICMSSN300", "ICMSSN400":
		result.ICMSSN102 = &ICMSSN102{
			Orig:  icms.Orig,
			CSOSN: icms.CST,
		}
	case "ICMSSN201":
		result.ICMSSN201 = &ICMSSN201{
			Orig:    icms.Orig,
			CSOSN:   icms.CST,
			ModBCST: *icms.ModBCST,
			PICMSST: *icms.PICMSST,
			VBCST:   *icms.VBCST,
			VICMSST: *icms.VICMSST,
		}
	case "ICMSSN202", "ICMSSN203":
		result.ICMSSN202 = &ICMSSN202{
			Orig:    icms.Orig,
			CSOSN:   icms.CST,
			ModBCST: *icms.ModBCST,
			PICMSST: *icms.PICMSST,
			VBCST:   *icms.VBCST,
			VICMSST: *icms.VICMSST,
		}
	case "ICMSSN500":
		result.ICMSSN500 = &ICMSSN500{
			Orig:       icms.Orig,
			CSOSN:      icms.CST,
			VBCSTRet:   *icms.VBCST,
			VICMSSTRet: *icms.VICMSST,
		}
	case "ICMSSN900":
		result.ICMSSN900 = &ICMSSN900{
			Orig:    icms.Orig,
			CSOSN:   icms.CST,
			ModBC:   *icms.ModBC,
			VBC:     *icms.VBC,
			PICMS:   *icms.PICMS,
			VICMS:   *icms.VICMS,
			ModBCST: *icms.ModBCST,
			PICMSST: *icms.PICMSST,
			VBCST:   *icms.VBCST,
			VICMSST: *icms.VICMSST,
		}
	}

	return result
}

// buildPIS builds PIS tax block
func (b *builder) buildPIS(pis PISInput) PIS {
	var result PIS

	switch pis.Tipo {
	case "PISAliq":
		result.PISAliq = &PISAliq{
			CST:  pis.CST,
			VBC:  *pis.VBC,
			PPIS: *pis.PPIS,
			VPIS: *pis.VPIS,
		}
	case "PISQtde":
		result.PISQtde = &PISQtde{
			CST:       pis.CST,
			QBCProd:   *pis.QBCProd,
			VAliqProd: *pis.VAliqProd,
			VPIS:      *pis.VPIS,
		}
	case "PISNT":
		result.PISNT = &PISNT{
			CST: pis.CST,
		}
	case "PISOutr":
		result.PISOutr = &PISOutr{
			CST:  pis.CST,
			VBC:  *pis.VBC,
			PPIS: *pis.PPIS,
			VPIS: *pis.VPIS,
		}
	}

	return result
}

// buildCOFINS builds COFINS tax block
func (b *builder) buildCOFINS(cofins COFINSInput) COFINS {
	var result COFINS

	switch cofins.Tipo {
	case "COFINSAliq":
		result.COFINSAliq = &COFINSAliq{
			CST:     cofins.CST,
			VBC:     *cofins.VBC,
			PCOFINS: *cofins.PCOFINS,
			VCOFINS: *cofins.VCOFINS,
		}
	case "COFINSQtde":
		result.COFINSQtde = &COFINSQtde{
			CST:       cofins.CST,
			QBCProd:   *cofins.QBCProd,
			VAliqProd: *cofins.VAliqProd,
			VCOFINS:   *cofins.VCOFINS,
		}
	case "COFINSNT":
		result.COFINSNT = &COFINSNT{
			CST: cofins.CST,
		}
	case "COFINSOutr":
		result.COFINSOutr = &COFINSOutr{
			CST:     cofins.CST,
			VBC:     *cofins.VBC,
			PCOFINS: *cofins.PCOFINS,
			VCOFINS: *cofins.VCOFINS,
		}
	}

	return result
}

// buildTotal builds total block
func (b *builder) buildTotal(itens []ItemInput) Total {
	var vBC, vICMS, vBCST, vST, vProd, vPIS, vCOFINS float64

	for _, item := range itens {
		vProdItem, _ := strconv.ParseFloat(item.VProd, 64)
		vProd += vProdItem

		// Calculate tax values based on ICMS
		if item.Imposto.ICMS.VBC != nil {
			vbc, _ := strconv.ParseFloat(*item.Imposto.ICMS.VBC, 64)
			vBC += vbc
		}
		if item.Imposto.ICMS.VICMS != nil {
			vicms, _ := strconv.ParseFloat(*item.Imposto.ICMS.VICMS, 64)
			vICMS += vicms
		}
		if item.Imposto.ICMS.VBCST != nil {
			vbcst, _ := strconv.ParseFloat(*item.Imposto.ICMS.VBCST, 64)
			vBCST += vbcst
		}
		if item.Imposto.ICMS.VICMSST != nil {
			vicmsst, _ := strconv.ParseFloat(*item.Imposto.ICMS.VICMSST, 64)
			vST += vicmsst
		}

		// PIS and COFINS
		if item.Imposto.PIS.VPIS != nil {
			vpis, _ := strconv.ParseFloat(*item.Imposto.PIS.VPIS, 64)
			vPIS += vpis
		}
		if item.Imposto.COFINS.VCOFINS != nil {
			vcofins, _ := strconv.ParseFloat(*item.Imposto.COFINS.VCOFINS, 64)
			vCOFINS += vcofins
		}
	}

	vNF := vProd + vST // Total value

	return Total{
		ICMSTot: ICMSTot{
			VBC:     fmt.Sprintf("%.2f", vBC),
			VICMS:   fmt.Sprintf("%.2f", vICMS),
			VBCST:   fmt.Sprintf("%.2f", vBCST),
			VST:     fmt.Sprintf("%.2f", vST),
			VProd:   fmt.Sprintf("%.2f", vProd),
			VPIS:    fmt.Sprintf("%.2f", vPIS),
			VCOFINS: fmt.Sprintf("%.2f", vCOFINS),
			VNF:     fmt.Sprintf("%.2f", vNF),
		},
	}
}

// buildTransp builds transport block
func (b *builder) buildTransp(transp TranspInput) Transp {
	return Transp{
		ModFrete: transp.ModFrete,
	}
}

// buildPag builds payment block
func (b *builder) buildPag(pagamentos []PagamentoInput) Pag {
	detPag := make([]DetPag, len(pagamentos))
	for i, pag := range pagamentos {
		detPag[i] = DetPag{
			TPag: pag.TPag,
			VPag: pag.VPag,
			Card: func() *Card {
				if pag.Card == nil {
					return nil
				}
				return &Card{
					TpIntegra: pag.Card.TpIntegra,
					CNPJ:      pag.Card.CNPJ,
					TBand:     pag.Card.TBand,
					CAut:      pag.Card.CAut,
				}
			}(),
		}
	}

	return Pag{
		DetPag: detPag,
		VTroco: nil, // Calculate if needed
	}
}

// buildInfIntermed builds intermediary information block
func (b *builder) buildInfIntermed(inf InfIntermedInput) InfIntermed {
	return InfIntermed{
		CNPJ:         inf.CNPJ,
		XNome:        inf.XNome,
		IdCadIntTran: inf.IdCadIntTran,
	}
}

// buildInfRespTec builds technical responsible information block
func (b *builder) buildInfRespTec(inf InfRespTecInput) InfRespTec {
	return InfRespTec{
		CNPJ:     inf.CNPJ,
		XContato: inf.XContato,
		Email:    inf.Email,
		Fone:     inf.Fone,
	}
}

// GenerateChaveAcesso generates the access key for NFC-e
func (b *builder) GenerateChaveAcesso(uf, cnpj, serie, nNF, tpEmis, cNF string, dhEmi time.Time) (string, error) {
	cUF := b.getCUF(uf)
	aamm := dhEmi.Format("0601") // YYMM

	// Clean CNPJ (remove non-numeric)
	cleanCNPJ := b.cleanNumericOnly(cnpj)
	if len(cleanCNPJ) != 14 {
		return "", fmt.Errorf("CNPJ deve ter 14 dígitos")
	}

	// Format: CUF + AAMM + CNPJ + MOD + SERIE + NNF + TPEMIS + CNF + DV
	chave := fmt.Sprintf("%02s%04s%014s65%03s%09s%01s%08s",
		cUF, aamm, cleanCNPJ, serie, nNF, tpEmis, cNF)

	dv := b.CalculateDV(chave)
	return chave + dv, nil
}

// CalculateDV calculates the check digit (DV) for the access key
func (b *builder) CalculateDV(chave string) string {
	if len(chave) != 43 {
		return "0" // Invalid length
	}

	// Weights for DV calculation (from right to left)
	weights := []int{2, 3, 4, 5, 6, 7, 8, 9}
	total := 0

	// Calculate weighted sum from right to left
	for i := 42; i >= 0; i-- {
		digit := int(chave[i] - '0')
		weight := weights[(42-i)%8]
		total += digit * weight
	}

	// Calculate DV
	remainder := total % 11
	if remainder == 0 || remainder == 1 {
		return "0"
	}
	return strconv.Itoa(11 - remainder)
}

// getCUF returns the federal unit code
func (b *builder) getCUF(uf string) string {
	ufCodes := map[string]string{
		"AC": "12", "AL": "27", "AP": "16", "AM": "13", "BA": "29",
		"CE": "23", "DF": "53", "ES": "32", "GO": "52", "MA": "21",
		"MT": "51", "MS": "50", "MG": "31", "PA": "15", "PB": "25",
		"PR": "41", "PE": "26", "PI": "22", "RJ": "33", "RN": "24",
		"RS": "43", "RO": "11", "RR": "14", "SC": "42", "SP": "35",
		"SE": "28", "TO": "17",
	}

	if code, exists := ufCodes[uf]; exists {
		return code
	}
	return "35" // Default to SP
}

// getMunicipioFG returns the municipality code for the federal unit
func (b *builder) getMunicipioFG(uf string) string {
	// Default municipalities for each UF (capital cities)
	municipios := map[string]string{
		"AC": "1200401", "AL": "2704302", "AP": "1600303", "AM": "1302603", "BA": "2927408",
		"CE": "2304400", "DF": "5300108", "ES": "3205309", "GO": "5208707", "MA": "2111300",
		"MT": "5103403", "MS": "5002704", "MG": "3106200", "PA": "1501402", "PB": "2507507",
		"PR": "4106902", "PE": "2611606", "PI": "2211001", "RJ": "3304557", "RN": "2408102",
		"RS": "4314902", "RO": "1100205", "RR": "1400100", "SC": "4205407", "SP": "3550308",
		"SE": "2800308", "TO": "1721000",
	}

	if code, exists := municipios[uf]; exists {
		return code
	}
	return "3550308" // Default to São Paulo
}

// cleanNumericOnly removes all non-numeric characters
func (b *builder) cleanNumericOnly(s string) string {
	var result []rune
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = append(result, r)
		}
	}
	return string(result)
}

// generateCNF generates a random 8-digit CNF (Código Numérico)
func (b *builder) generateCNF() string {
	// Generate random 8-digit number (00000001 to 99999999)
	// In production, ensure uniqueness within the company for the day
	return fmt.Sprintf("%08d", time.Now().UnixNano()%99999999+1)
}
