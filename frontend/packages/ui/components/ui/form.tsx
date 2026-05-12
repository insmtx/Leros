"use client";

import { cn } from "@leros/ui/lib/utils";
import type * as React from "react";

function Form({ className, ...props }: React.ComponentProps<"form">) {
	return <form data-slot="form" className={cn(className)} {...props} />;
}

export { Form };
