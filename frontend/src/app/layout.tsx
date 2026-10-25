import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import Providers from "@/components/Providers";
import { themeInitScript } from "@/lib/theme";

const inter = Inter({ subsets: ["latin"] });
const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono", preload: false });

export const metadata: Metadata = {
  title: "WebChat",
  description: "Real-time chat rooms",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* Constant script (no user data): runs before first paint to avoid a theme flash. */}
        {/* eslint-disable-next-line react/no-danger */}
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className={`${inter.className} ${mono.variable}`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}