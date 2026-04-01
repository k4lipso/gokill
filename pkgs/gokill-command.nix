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
  pname = "gokill";
  version = "0.5";
  vendorHash = vendorHash;
  src = fs.toSource {
    root = ../.;
    fileset = sourceFiles;
  };
  buildInputs = [
    pkgs.olm
  ];

  postInstall = ''
    cp -r ./etc $out/ #for .deb packages
    '';
}


