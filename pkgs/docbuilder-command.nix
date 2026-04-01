{ pkgs, ... }:

let
  fs = pkgs.lib.fileset;
  sourceFiles =  fs.difference
    (fs.gitTracked ../.)
    (fs.unions [
      (fs.maybeMissing ../result)
      (fs.fileFilter (file: file.hasExt "nix") ../.)
      (fs.fileFilter (file: file.hasExt "md") ../.)
      (fs.fileFilter (file: file.hasExt "json") ../.)
    ]);
  vendorHash = (import ./vendorhash.nix).vendorHash;
in
pkgs.buildGoModule {
  pname = "docbuilder";
  version = "1.0";
  vendorHash = vendorHash;
  buildFLags = "-o . $dest/cmd/gokill/docbuilder";
  src = fs.toSource {
    root = ../.;
    fileset = sourceFiles;
  };

  buildInputs = [
    pkgs.olm
  ];

  postInstall = ''
    '';
}


