import { ThemeProvider } from "@singeros/ui/components/common/theme-provider";
import { Toaster } from "@singeros/ui/components/ui/sonner";
import { cn } from "@singeros/ui/lib/utils";
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
	subsets: ["latin"],
	variable: "--font-sans",
});

export const metadata: Metadata = {
	title: "SingerOS",
	description: "AI OS for Software Engineering",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html
			lang="zh-CN"
			suppressHydrationWarning
			className={cn("antialiased font-sans h-full", inter.variable)}
		>
			<body className="h-full">
				<ThemeProvider defaultTheme="system">
					{children}
					<Toaster />
				</ThemeProvider>
			</body>
		</html>
	);
}
