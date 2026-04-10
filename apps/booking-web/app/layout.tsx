import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Book an Appointment — Zenvikar",
  description: "Book an appointment with your favorite service provider",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
