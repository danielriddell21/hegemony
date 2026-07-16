package cli

import (
	"github.com/spf13/cobra"

	"github.com/danielriddell21/hegemony/internal/sim"
)

type paramFlags struct {
	income         int
	neutralDefense int
	startStrength  int
	maxStrength    int
	jitter         float64
}

func addParamFlags(cmd *cobra.Command) *paramFlags {
	d := sim.DefaultParams()
	pf := &paramFlags{}
	cmd.Flags().IntVar(&pf.income, "income", d.IncomePerCell, "strength each owned cell gains per tick")
	cmd.Flags().IntVar(&pf.neutralDefense, "neutral-defense", d.NeutralDefense, "strength neutral cells defend with")
	cmd.Flags().IntVar(&pf.startStrength, "start-strength", d.StartStrength, "strength of each faction's spawn cell")
	cmd.Flags().IntVar(&pf.maxStrength, "max-strength", d.MaxCellStrength, "maximum strength a single cell can hold")
	cmd.Flags().Float64Var(&pf.jitter, "jitter", d.Jitter, "combat randomness in [0,1)")
	return pf
}

func (pf *paramFlags) params() sim.Params {
	return sim.Params{
		IncomePerCell:   pf.income,
		NeutralDefense:  pf.neutralDefense,
		StartStrength:   pf.startStrength,
		MaxCellStrength: pf.maxStrength,
		Jitter:          pf.jitter,
	}
}
