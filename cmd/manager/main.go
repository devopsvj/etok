package manager

import (
	"flag"
	"fmt"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/leg100/stok/api/command"
	"github.com/leg100/stok/api/v1alpha1"
	"github.com/leg100/stok/controllers"
	"github.com/leg100/stok/scheme"
	"github.com/leg100/stok/version"

	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = ctrl.Log.WithName("setup")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

type operatorCmd struct {
	metricsAddr          string
	enableLeaderElection bool

	cmd *cobra.Command
}

func NewOperatorCmd() *cobra.Command {
	c := &operatorCmd{}
	c.cmd = &cobra.Command{
		Use:   "operator",
		Short: "Run the stok operator",
		Args:  cobra.NoArgs,
		RunE:  c.doOperatorCmd,
	}
	c.cmd.Flags().StringVar(&c.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	c.cmd.Flags().BoolVar(&c.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	c.cmd.Flags().AddGoFlagSet(flag.CommandLine)

	return c.cmd
}

func (c *operatorCmd) doOperatorCmd(cmd *cobra.Command, args []string) error {
	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	printVersion()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: c.metricsAddr,
		Port:               9443,
		LeaderElection:     c.enableLeaderElection,
		LeaderElectionID:   "688c905b.goalspike.com",
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// Setup workspace ctrl with mgr
	if err = (&controllers.WorkspaceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Workspace"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create workspace controller: %w", err)
	}

	// Setup each command ctrl with mgr
	for _, kind := range command.CommandKinds {
		o, err := scheme.Scheme.New(v1alpha1.SchemeGroupVersion.WithKind(kind))
		if err != nil {
			return err
		}

		if err = (&controllers.CommandReconciler{
			Client: mgr.GetClient(),
			// TODO: rename to c to something less silly, like cmd
			C:            o.(command.Interface),
			Log:          ctrl.Log.WithName("controllers").WithName(kind),
			Kind:         kind,
			ResourceType: command.CommandKindToType(kind),
			Scheme:       mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create %s controller: %w", command.CommandKindToCLI(kind), err)
		}

	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}
