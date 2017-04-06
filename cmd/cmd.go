			Ui:                meta.Ui,
			}, nil
		},
	
               "status": func() (cli.Command, error) {


			return &cmd.StatusCommand{
			// The folowing lines are for advanced
				/*Status: sts,
				IP:     ip,    
				Port:   port,*/
				Ui: meta.Ui,
			}, nil
		},
 	}
