$ErrorActionPreference = 'Stop'

$root = 'C:\dev\home-telemetry\server'
$certDir = Join-Path $root 'certs'
New-Item -ItemType Directory -Force -Path $certDir | Out-Null

$certPath = Join-Path $certDir 'localhost-cert.pem'
$keyPath = Join-Path $certDir 'localhost-key.pem'

$dnsName = 'localhost'
$notAfter = (Get-Date).AddYears(2)

$rsa = New-Object System.Security.Cryptography.RSACryptoServiceProvider(2048)
$hashAlg = [System.Security.Cryptography.HashAlgorithmName]::SHA256
$padding = [System.Security.Cryptography.RSASignaturePadding]::Pkcs1

$req = New-Object System.Security.Cryptography.X509Certificates.CertificateRequest(
    "CN=$dnsName",
    $rsa,
    $hashAlg,
    $padding
)

$sanBuilder = New-Object System.Security.Cryptography.X509Certificates.SubjectAlternativeNameBuilder
$sanBuilder.AddDnsName($dnsName)
$req.CertificateExtensions.Add($sanBuilder.Build())

$basicConstraints = New-Object System.Security.Cryptography.X509Certificates.X509BasicConstraintsExtension($false, $false, 0, $false)
$ds = [int][System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::DigitalSignature
$ke = [int][System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]::KeyEncipherment
$keyUsageFlags = [System.Security.Cryptography.X509Certificates.X509KeyUsageFlags]($ds -bor $ke)
$keyUsage = New-Object System.Security.Cryptography.X509Certificates.X509KeyUsageExtension($keyUsageFlags, $false)
$ekuOids = New-Object System.Security.Cryptography.OidCollection
$ekuOids.Add((New-Object System.Security.Cryptography.Oid('1.3.6.1.5.5.7.3.1'))) | Out-Null
$eku = New-Object System.Security.Cryptography.X509Certificates.X509EnhancedKeyUsageExtension($ekuOids, $false)

$req.CertificateExtensions.Add($basicConstraints)
$req.CertificateExtensions.Add($keyUsage)
$req.CertificateExtensions.Add($eku)

$cert = $req.CreateSelfSigned((Get-Date).AddDays(-1), $notAfter)

function Write-Pem {
    param(
        [Parameter(Mandatory)] [string] $Path,
        [Parameter(Mandatory)] [string] $Label,
        [Parameter(Mandatory)] [byte[]] $Bytes
    )
    $base64 = [System.Convert]::ToBase64String($Bytes)
    $lines = $base64 -split "(.{1,64})" | Where-Object { $_ -ne "" }
    $content = @(
        "-----BEGIN $Label-----"
    ) + $lines + @(
        "-----END $Label-----"
    )
    Set-Content -Encoding ascii -Path $Path -Value $content
}

function Encode-Length([int] $len) {
    if ($len -lt 128) {
        return ,([byte]$len)
    }
    $bytes = @()
    while ($len -gt 0) {
        $bytes = ,([byte]($len -band 0xFF)) + $bytes
        $len = $len -shr 8
    }
    return ,([byte](0x80 -bor $bytes.Length)) + $bytes
}

function Encode-Integer([byte[]] $value) {
    if ($value.Length -eq 0) {
        $value = ,0
    }
    if ($value[0] -band 0x80) {
        $value = ,0 + $value
    }
    return ,0x02 + (Encode-Length $value.Length) + $value
}

function Encode-Sequence([byte[]] $value) {
    return ,0x30 + (Encode-Length $value.Length) + $value
}

$params = $rsa.ExportParameters($true)

$version = Encode-Integer (,0)
$n = Encode-Integer $params.Modulus
$e = Encode-Integer $params.Exponent
$d = Encode-Integer $params.D
$p = Encode-Integer $params.P
$q = Encode-Integer $params.Q
$dp = Encode-Integer $params.DP
$dq = Encode-Integer $params.DQ
$iq = Encode-Integer $params.InverseQ

$seq = @()
$seq += $version + $n + $e + $d + $p + $q + $dp + $dq + $iq
$pkcs1 = Encode-Sequence $seq

$certBytes = $cert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Cert)
Write-Pem -Path $certPath -Label 'CERTIFICATE' -Bytes $certBytes

Write-Pem -Path $keyPath -Label 'RSA PRIVATE KEY' -Bytes $pkcs1

Write-Host "Wrote $certPath"
Write-Host "Wrote $keyPath"
