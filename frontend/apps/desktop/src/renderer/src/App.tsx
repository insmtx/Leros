import { ThemeProvider } from "@leros/ui/components/common/theme-provider";
import { Toaster } from "@leros/ui/components/ui/sonner";
import { BrowserRouter } from "react-router-dom";
import { AppRoutes } from "./routes";

export default function App() {
	return (
		<BrowserRouter>
			<ThemeProvider defaultTheme="system">
				<AppRoutes />
				<Toaster />
			</ThemeProvider>
		</BrowserRouter>
	);
}
