// Copyright 2023, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { useLongClick } from "@/app/hook/useLongClick";
import { makeIconClass } from "@/util/util";
import clsx from "clsx";
import { memo, useRef } from "react";
import "./iconbutton.scss";

export const IconButton = memo(({ decl, className }: { decl: IconButtonDecl; className?: string }) => {
    const buttonRef = useRef<HTMLDivElement>(null);
    useLongClick(buttonRef, decl.click, decl.longClick, decl.disabled);
    return (
        <div
            ref={buttonRef}
            className={clsx("iconbutton", className, decl.className, { disabled: decl.disabled })}
            title={decl.title}
            style={{ color: decl.iconColor ?? "inherit" }}
        >
            {typeof decl.icon === "string" ? <i className={makeIconClass(decl.icon, true)} /> : decl.icon}
        </div>
    );
});
