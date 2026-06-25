# ──────────────────────────────────────────────────────────────────────
# SES — Domain verification + SMTP IAM user for agent-registration OTP
# email delivery.  DNS records created in Cloudflare; SMTP IAM user
# created here, credential conversion happens in CI (needs aws CLI).
# ──────────────────────────────────────────────────────────────────────

# Domain identity — verifies that we own opennhp.org for SES sending.
resource "aws_ses_domain_identity" "opennhp" {
  domain = "opennhp.org"
}

# SES domain verification TXT record in Cloudflare.
# The verification token goes in a TXT record at _amazones.opennhp.org.
resource "cloudflare_record" "ses_verification" {
  zone_id = var.cloudflare_zone_id
  name    = "_amazones"
  content = aws_ses_domain_identity.opennhp.verification_token
  type    = "TXT"
  ttl     = 300
  comment = "SES domain verification — managed by Terraform"
}

# Enable DKIM signing on the verified domain.
resource "aws_ses_domain_dkim" "opennhp" {
  domain = aws_ses_domain_identity.opennhp.domain
}

# DKIM records (3 CNAME entries).  Tokens come from aws_ses_domain_dkim.
resource "cloudflare_record" "ses_dkim" {
  count   = 3
  zone_id = var.cloudflare_zone_id
  name    = "${aws_ses_domain_dkim.opennhp.dkim_tokens[count.index]}._domainkey"
  content = "${aws_ses_domain_dkim.opennhp.dkim_tokens[count.index]}.dkim.amazonses.com"
  type    = "CNAME"
  ttl     = 300
  comment = "SES DKIM ${count.index + 1}/3 — managed by Terraform"
}

# MAIL FROM subdomain for bounce/complaint handling.
resource "aws_ses_domain_mail_from" "opennhp" {
  domain           = aws_ses_domain_identity.opennhp.domain
  mail_from_domain = "mail.opennhp.org"
}

# MX record for the MAIL FROM subdomain.
resource "cloudflare_record" "ses_mail_from_mx" {
  zone_id = var.cloudflare_zone_id
  name    = aws_ses_domain_mail_from.opennhp.mail_from_domain
  content = "feedback-smtp.${var.aws_region}.amazonses.com"
  type    = "MX"
  ttl     = 300
  priority = 10
  comment = "SES MAIL FROM MX — managed by Terraform"
}

# SPF record for the MAIL FROM subdomain.
resource "cloudflare_record" "ses_mail_from_txt" {
  zone_id = var.cloudflare_zone_id
  name    = aws_ses_domain_mail_from.opennhp.mail_from_domain
  content = "v=spf1 include:amazonses.com ~all"
  type    = "TXT"
  ttl     = 300
  comment = "SES MAIL FROM SPF — managed by Terraform"
}

# Verify the noreply@ address so the demo can send immediately
# (required while the account is in the SES sandbox).
resource "aws_ses_email_identity" "noreply" {
  email = "noreply@opennhp.org"
}

# ── SMTP IAM user ───────────────────────────────────────────────────
# SES SMTP uses IAM credentials.  The IAM user gets ses:SendRawEmail;
# the CI pipeline derives the SMTP password from the IAM access key
# secret via:  aws ses get-smtp-password --secret-key <key> --region us-east-2

resource "aws_iam_user" "ses_smtp" {
  name = "opennhp-demo-ses-smtp"
  tags = var.tags
}

resource "aws_iam_user_policy" "ses_smtp" {
  name = "ses-send-email"
  user = aws_iam_user.ses_smtp.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ses:SendEmail",
          "ses:SendRawEmail",
        ]
        Resource = ["*"]
        Condition = {
          StringEquals = {
            "ses:FromAddress" = "noreply@opennhp.org"
          }
        }
      },
      {
        Effect   = "Allow"
        Action   = ["ses:GetSendQuota"]
        Resource = ["*"]
      }
    ]
  })
}

resource "aws_iam_access_key" "ses_smtp" {
  user = aws_iam_user.ses_smtp.name
}
