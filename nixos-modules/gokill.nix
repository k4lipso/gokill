{ config, lib, pkgs, inputs, ... }: 
let
  cfg = config.services.gokill;
  configFile = pkgs.writeText "config.json" (builtins.toJSON cfg.triggers); 
  gokill-pkg = inputs.gokill.packages.aarch64-linux.gokill;
  testRun = if cfg.testRun then "-t" else "";
in
{
  options = with lib; {
    services.gokill = {
      enable = mkOption {
        default = false;
        type = types.bool;
        description = mdDoc ''
          Enables gokill daemon
          '';
      };

      testRun = mkOption {
        default = false;
        type = types.bool;
        description = mdDoc ''
          When set to true gokill is performing a test run
          '';
      };

      triggers = mkOption {
        description = "list of triggers";
        default = [];
        type = with types; types.listOf ( submodule {
          options = {
            type = mkOption {
              type = types.str;
            };

            name = mkOption {
              type = types.str;
            };

            options = mkOption {
              type = types.attrs;
            };

            actions = mkOption {
              description = "list of actions";
              type = with types; types.listOf ( submodule {
                options = {
                  type = mkOption {
                    type = types.str;
                  };

                  options = mkOption {
                    type = types.attrs;
                  };

                  stage = mkOption {
                    type = types.int;
                  };
                };
              });
            };
          };
        });
      };
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.setup-gokill = {
      description = "Initialize gokill directory";
      wantedBy = [ "gokill.service" ];
      unitConfig.ConditionPathExists = "!/etc/gokill/.is_initialized";
      serviceConfig = {
        Type = "oneshot";
      };
      script = ''
        mkdir /etc/gokill
        touch /etc/gokill/.is_initialized
      '';
    };

    systemd.services.gokill = {
      description = "gokill daemon";
      serviceConfig = {
        Type = "simple";
        ExecStart = "${gokill-pkg}/bin/gokill --db /etc/gokill -r -c ${configFile} ${testRun}";
        Restart = "on-failure";
      };

      wantedBy = [ "default.target" ];
    };
  };
}


